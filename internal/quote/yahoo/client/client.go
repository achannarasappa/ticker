package client

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
)

const (
	userAgent                             = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	userAgentClientHintBrandingAndVersion = "\"Google Chrome\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24\""
	userAgentClientHintPlatform           = "\"Windows\""
)

var errSessionRefresh = errors.New("yahoo session refresh error")

func New(clientMain *resty.Client, clientSession *resty.Client) *resty.Client {

	client := clientMain.
		SetBaseURL("https://query1.finance.yahoo.com").
		SetHeader("authority", "query1.finance.yahoo.com").
		SetHeader("accept", "*/*").
		SetHeader("accept-language", "en-US,en;q=0.9,ja;q=0.8").
		SetHeader("origin", "https://finance.yahoo.com").
		SetHeader("sec-ch-ua", userAgentClientHintBrandingAndVersion).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", userAgentClientHintPlatform).
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "same-site").
		SetHeader("user-agent", userAgent).
		SetQueryParam("formatted", "true").
		SetQueryParam("lang", "en-US").
		SetQueryParam("region", "US").
		SetQueryParam("corsDomain", "finance.yahoo.com").
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(1)).
		AddRetryAfterErrorCondition().
		SetRetryCount(1).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {

			if r.IsError() {
				return RefreshSession(c, clientSession)
			}

			return nil
		})

	return client

}

func RefreshSession(clientMain *resty.Client, clientSession *resty.Client) error {
	var err error
	var cookies []*http.Cookie
	var crumb string

	cookies, err = getCookie(clientSession)

	if err != nil {
		return err
	}

	crumb, err = getCrumb(clientSession, cookies)

	if err != nil {
		return err
	}

	clientMain.
		SetCookies(cookies).
		SetQueryParam("crumb", crumb)

	return nil
}

func getCookie(client *resty.Client) ([]*http.Cookie, error) {

	res, err := client.
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(1)).
		R().
		SetHeader("authority", "finance.yahoo.com").
		SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7").
		SetHeader("accept-language", "en-US,en;q=0.9").
		SetHeader("sec-ch-ua", userAgentClientHintBrandingAndVersion).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", userAgentClientHintPlatform).
		SetHeader("sec-fetch-dest", "document").
		SetHeader("sec-fetch-mode", "navigate").
		SetHeader("sec-fetch-site", "none").
		SetHeader("sec-fetch-user", "?1").
		SetHeader("upgrade-insecure-requests", "1").
		SetHeader("user-agent", userAgent).
		Get("https://finance.yahoo.com/")

	if err != nil && !strings.Contains(err.Error(), "stopped after") {
		return nil, fmt.Errorf("error requesting a cookie: %w", err)
	}

	if isEUConsentRedirect(res) {
		return getCookieEU(client)
	}

	x := res.Cookies()
	if !isRequiredCookieSet(res) {
		return nil, fmt.Errorf("%w: unexpected response from Yahoo API: A3 session cookie missing from response", errSessionRefresh)
	}

	return x, nil

}

func getCookieEU(client *resty.Client) ([]*http.Cookie, error) {

	var cookies []*http.Cookie

	reCsrfToken := regexp.MustCompile("gcrumb=(?:([A-Za-z0-9_]*))")
	reSessionID := regexp.MustCompile("sessionId=(?:([A-Za-z0-9_-]*))")

	res1, err1 := client.
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(3)).
		R().
		SetHeader("authority", "finance.yahoo.com").
		SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7").
		SetHeader("accept-language", "en-US,en;q=0.9").
		SetHeader("sec-ch-ua", userAgentClientHintBrandingAndVersion).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", userAgentClientHintPlatform).
		SetHeader("sec-fetch-dest", "document").
		SetHeader("sec-fetch-mode", "navigate").
		SetHeader("sec-fetch-site", "none").
		SetHeader("sec-fetch-user", "?1").
		SetHeader("upgrade-insecure-requests", "1").
		SetHeader("user-agent", userAgent).
		Get("https://finance.yahoo.com/")

	if err1 != nil && !strings.Contains(err1.Error(), "stopped after") {
		return cookies, fmt.Errorf("error attempting to get Yahoo API session id: %w", err1)
	}

	if !strings.HasPrefix(res1.Status(), "2") {
		return cookies, fmt.Errorf("%w: unexpected response from Yahoo API: non-2xx response code: %s", errSessionRefresh, res1.Status())
	}

	sessionIDMatchResult := reSessionID.FindStringSubmatch(res1.RawResponse.Request.URL.String())

	if len(sessionIDMatchResult) != 2 {
		return cookies, fmt.Errorf("%w: error unable to extract session id from redirected request URL: %s", errSessionRefresh, res1.Request.URL)
	}

	sessionID := sessionIDMatchResult[1]

	csrfTokenMatchResult := reCsrfToken.FindStringSubmatch(res1.RawResponse.Request.Response.Request.URL.String())

	if len(csrfTokenMatchResult) != 2 {
		return cookies, fmt.Errorf("%w: error unable to extract CSRF token from Location header: '%s'", errSessionRefresh, res1.Header().Get("Location"))
	}

	csrfToken := csrfTokenMatchResult[1]

	GUCSCookie := res1.RawResponse.Request.Response.Request.Response.Cookies()

	if len(GUCSCookie) == 0 {
		return cookies, fmt.Errorf("%w: no cookies set by finance.yahoo.com", errSessionRefresh)
	}

	res2, err2 := client.
		SetRedirectPolicy(resty.FlexibleRedirectPolicy(2)).
		SetContentLength(true).
		R().
		SetHeader("origin", "https://consent.yahoo.com").
		SetHeader("host", "consent.yahoo.com").
		SetHeader("content-type", "application/x-www-form-urlencoded").
		SetHeader("accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8").
		SetHeader("accept-language", "en-US,en;q=0.5").
		SetHeader("accept-encoding", "gzip, deflate, br").
		SetHeader("dnt", "1").
		SetHeader("sec-ch-ua", userAgentClientHintBrandingAndVersion).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", userAgentClientHintPlatform).
		SetHeader("sec-fetch-dest", "document").
		SetHeader("sec-fetch-mode", "navigate").
		SetHeader("sec-fetch-site", "same-origin").
		SetHeader("sec-fetch-user", "?1").
		SetHeader("referer", "https://consent.yahoo.com/v2/collectConsent?sessionId="+sessionID).
		SetHeader("user-agent", userAgent).
		SetCookies(GUCSCookie).
		SetFormData(map[string]string{
			"csrfToken": csrfToken,
			"sessionId": sessionID,
			"namespace": "yahoo",
			"agree":     "agree",
		}).
		Post("https://consent.yahoo.com/v2/collectConsent?sessionId=" + sessionID)

	if err2 != nil && !strings.Contains(err2.Error(), "stopped after") {
		return cookies, fmt.Errorf("error attempting to agree to EU consent request: %w", err2)
	}

	if !isRequiredCookieSet(res2) {
		return nil, fmt.Errorf("%w: unexpected response from Yahoo API: A3 session cookie missing from response after agreeing to EU consent request: %s", errSessionRefresh, res2.Status())
	}

	return res2.Cookies(), nil

}

func getCrumb(client *resty.Client, cookies []*http.Cookie) (string, error) {
	res, err := client.R().
		SetHeader("authority", "query2.finance.yahoo.com").
		SetHeader("accept", "*/*").
		SetHeader("accept-language", "en-US,en;q=0.9,ja;q=0.8").
		SetHeader("content-type", "text/plain").
		SetHeader("origin", "https://finance.yahoo.com").
		SetHeader("sec-ch-ua", userAgentClientHintBrandingAndVersion).
		SetHeader("sec-ch-ua-mobile", "?0").
		SetHeader("sec-ch-ua-platform", userAgentClientHintPlatform).
		SetHeader("sec-fetch-dest", "empty").
		SetHeader("sec-fetch-mode", "cors").
		SetHeader("sec-fetch-site", "same-site").
		SetHeader("user-agent", userAgent).
		SetCookies(cookies).
		Get("https://query2.finance.yahoo.com/v1/test/getcrumb")

	if err != nil {
		return "", fmt.Errorf("error requesting a crumb: %w", err)
	}

	if !strings.HasPrefix(res.Status(), "2") {
		return "", fmt.Errorf("%w: unexpected response from Yahoo API when attempting to retrieve crumb: non-2xx response code: %s", errSessionRefresh, res.Status())
	}

	return res.String(), err
}

func isRequiredCookieSet(res *resty.Response) bool {

	cookies := res.Cookies()

	for _, cookie := range cookies {
		if cookie.Name == "A3" {
			return true
		}
	}

	return false

}

func isEUConsentRedirect(res *resty.Response) bool {
	return strings.Contains(res.Header().Get("Location"), "guce.yahoo.com") &&
		strings.HasPrefix(res.Status(), "3")
}

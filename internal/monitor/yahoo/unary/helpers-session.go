package unary

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

// Constants for URLs and common header values
const (
	sessionCrumbPath          = "/v1/test/getcrumb"
	sessionConsentPathPattern = "/v2/collectConsent?sessionId=%s"

	// Common header values
	defaultUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/134.0.0.0 Safari/537.36"
	defaultAcceptValue = "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"
	defaultAcceptLang  = "en-US,en;q=0.9"
)

// refreshSession refreshes the Yahoo Finance session by getting new cookies and crumb
func (u *UnaryAPI) refreshSession() error {
	var err error

	// Get cookies
	u.cookies, err = u.getCookie()
	if err != nil {
		return err
	}

	// Get crumb
	u.crumb, err = u.getCrumb()
	if err != nil {
		return err
	}

	return nil
}

// getCookie retrieves authentication cookies from Yahoo Finance
func (u *UnaryAPI) getCookie() ([]*http.Cookie, error) {
	req, err := http.NewRequest(http.MethodGet, u.sessionRootURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating cookie request: %w", err)
	}

	req.Header.Set("Authority", "finance.yahoo.com")
	req.Header.Set("Accept", defaultAcceptValue)
	req.Header.Set("Accept-Language", defaultAcceptLang)
	req.Header.Set("User-Agent", defaultUserAgent)

	resp, err := u.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error requesting a cookie: %w", err)
	}
	defer resp.Body.Close()

	// Check for EU consent redirect
	if isEUConsentRedirect(resp) {
		return u.getCookieEU()
	}

	cookies := resp.Cookies()
	if !isRequiredCookieSet(cookies) {
		return nil, errors.New("session refresh error: A3 session cookie missing from response")
	}

	return cookies, nil
}

// getCookieEU handles the EU consent flow to get cookies
func (u *UnaryAPI) getCookieEU() ([]*http.Cookie, error) {
	var cookies []*http.Cookie

	// Create a client with a redirect limit of 3 instead of the default of 1
	client1 := u.createClientWithRedirectLimit(3)

	// First request to get redirected to consent page
	req1, _ := http.NewRequest(http.MethodGet, u.sessionRootURL, nil)

	req1.Header.Set("Authority", "finance.yahoo.com")
	req1.Header.Set("Accept", defaultAcceptValue)
	req1.Header.Set("Accept-Language", defaultAcceptLang)
	req1.Header.Set("User-Agent", defaultUserAgent)

	resp1, err := client1.Do(req1)
	if err != nil {
		return nil, fmt.Errorf("error refreshing EU session: %w", err)
	}
	defer resp1.Body.Close()

	if resp1.StatusCode < 200 || resp1.StatusCode >= 300 {
		return nil, fmt.Errorf("session refresh error: unexpected response from Yahoo API: non-2xx response code: %d", resp1.StatusCode)
	}

	// Extract session ID and CSRF token from URL
	sessionID, csrfToken, err := extractSessionAndCSRF(resp1)
	if err != nil {
		return nil, err
	}

	// Get GUCS cookie
	gucsCookies := resp1.Cookies()
	if len(gucsCookies) == 0 {
		return nil, errors.New("session refresh error: GUCS cookie missing from response")
	}

	// Submit consent form
	formData := url.Values{}
	formData.Set("csrfToken", csrfToken)
	formData.Set("sessionId", sessionID)
	formData.Set("namespace", "yahoo")
	formData.Set("agree", "agree")

	formDataStr := formData.Encode()

	req2, err := http.NewRequest(http.MethodPost, u.sessionConsentURL+fmt.Sprintf(sessionConsentPathPattern, sessionID), strings.NewReader(formDataStr))
	if err != nil {
		return nil, fmt.Errorf("error creating consent submission request: %w", err)
	}

	req2.Header.Set("Origin", "https://consent.yahoo.com")
	req2.Header.Set("Host", "consent.yahoo.com")
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req2.Header.Set("Accept", defaultAcceptValue)
	req2.Header.Set("Accept-Language", defaultAcceptLang)
	req2.Header.Set("Referer", u.sessionConsentURL+fmt.Sprintf(sessionConsentPathPattern, sessionID))
	req2.Header.Set("User-Agent", defaultUserAgent)
	req2.Header.Set("Content-Length", strconv.Itoa(len(formDataStr)))
	req2.Header.Set("Accept-Encoding", "gzip, deflate, br")
	req2.Header.Set("DNT", "1")
	req2.Header.Set("Sec-Ch-Ua", "\"Google Chrome\";v=\"134\", \"Chromium\";v=\"134\", \"Not-A.Brand\";v=\"24\"")
	req2.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req2.Header.Set("Sec-Ch-Ua-Platform", "\"Windows\"")
	req2.Header.Set("Sec-Fetch-Dest", "document")
	req2.Header.Set("Sec-Fetch-Mode", "navigate")
	req2.Header.Set("Sec-Fetch-Site", "same-origin")
	req2.Header.Set("Sec-Fetch-User", "?1")

	// Add GUCS cookies
	for _, cookie := range gucsCookies {
		req2.AddCookie(cookie)
	}

	client2 := u.createClientWithRedirectLimit(2)

	resp2, err := client2.Do(req2)
	if err != nil {
		return nil, fmt.Errorf("HTTP protocol error attempting to agree to EU consent request: %w", err)
	}
	defer resp2.Body.Close()

	cookies = resp2.Cookies()
	if !isRequiredCookieSet(cookies) {
		return nil, errors.New("session refresh error: A3 session cookie missing from response after agreeing to EU consent request")
	}

	return cookies, nil
}

// getCrumb retrieves the crumb value needed for authenticated requests
func (u *UnaryAPI) getCrumb() (string, error) {
	req, err := http.NewRequest(http.MethodGet, u.sessionCrumbURL+sessionCrumbPath, nil)
	if err != nil {
		return "", fmt.Errorf("error creating crumb request: %w", err)
	}

	req.Header.Set("Authority", "query2.finance.yahoo.com")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", defaultAcceptLang)
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Origin", u.sessionRootURL)
	req.Header.Set("User-Agent", defaultUserAgent)

	// Add cookies
	for _, cookie := range u.cookies {
		req.AddCookie(cookie)
	}

	resp, err := u.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error requesting a crumb: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("session refresh error: unexpected response from Yahoo API when attempting to retrieve crumb: non-2xx response code: %d", resp.StatusCode)
	}

	// Read crumb from response body
	crumbBytes, _ := io.ReadAll(resp.Body)
	crumb := string(crumbBytes)

	return crumb, nil
}

// createClientWithRedirectLimit returns a new http.Client with the specified redirect limit
func (a *UnaryAPI) createClientWithRedirectLimit(limit int) *http.Client {
	return &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= limit {
				return http.ErrUseLastResponse
			}

			return nil
		},
	}
}

// isEUConsentRedirect checks if the response is a redirect to the EU consent page
func isEUConsentRedirect(resp *http.Response) bool {
	return strings.Contains(resp.Header.Get("Location"), "/consent") &&
		resp.StatusCode >= 300 && resp.StatusCode < 400
}

// isRequiredCookieSet checks if the A3 cookie is present in the cookies
func isRequiredCookieSet(cookies []*http.Cookie) bool {
	for _, cookie := range cookies {
		if cookie.Name == "A3" {
			return true
		}
	}

	return false
}

// extractSessionAndCSRF extracts session ID and CSRF token from response
func extractSessionAndCSRF(resp *http.Response) (string, string, error) {
	// Extract session ID from URL
	sessionIDRegex := regexp.MustCompile("sessionId=(?:([A-Za-z0-9_-]*))")
	csrfTokenRegex := regexp.MustCompile("gcrumb=(?:([A-Za-z0-9_]*))")

	// Check for session ID in Location header or URL
	var sessionIDMatch []string
	var csrfTokenMatch []string

	if resp.Request.URL != nil {
		sessionIDMatch = sessionIDRegex.FindStringSubmatch(resp.Request.URL.String())
	}

	if len(sessionIDMatch) < 2 {
		return "", "", errors.New("session refresh error: error unable to extract session id from redirected request URL")
	}

	// Check for CSRF token in URL
	if resp.Request.Response != nil && resp.Request.Response.Request != nil && resp.Request.Response.Request.URL != nil {
		csrfTokenMatch = csrfTokenRegex.FindStringSubmatch(resp.Request.Response.Request.URL.String())
	}

	if len(csrfTokenMatch) < 2 {
		return "", "", errors.New("session refresh error: error unable to extract CSRF token from Location header")
	}

	return sessionIDMatch[1], csrfTokenMatch[1], nil
}

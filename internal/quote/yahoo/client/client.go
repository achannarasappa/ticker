package client

import (
	"net/http"

	"github.com/go-resty/resty/v2"
)

const (
	userAgent                             = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/113.0.0.0 Safari/537.36"
	userAgentClientHintBrandingAndVersion = "\"Google Chrome\";v=\"113\", \"Chromium\";v=\"113\", \"Not-A.Brand\";v=\"24\""
	userAgentClientHintPlatform           = "\"Windows\""
)

func New() *resty.Client {

	client := resty.New().
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
		AddRetryAfterErrorCondition().
		SetRetryCount(1).
		OnAfterResponse(func(c *resty.Client, r *resty.Response) error {

			if r.IsError() {
				refreshClient(c)
			}

			return nil
		})

	return client

}

func refreshClient(c *resty.Client) {
	cookies := getCookie(c)
	crumb := getCrumb(c, cookies)

	c.
		SetCookies(cookies).
		SetQueryParam("crumb", crumb)
}

func getCookie(client *resty.Client) []*http.Cookie {

	res, _ := client.R().
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

	return res.Cookies()
}

func getCrumb(client *resty.Client, cookies []*http.Cookie) string {
	res, _ := client.R().
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

	return res.String()
}

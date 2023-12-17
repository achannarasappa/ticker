package client_test

import (
	"errors"
	"net/http"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/quote/yahoo/client"
)

var _ = Describe("Client", func() {

	Describe("New", func() {

		It("should return a new HTTP client", func() {

			client := c.New(clientResty, resty.New())

			Expect(client.BaseURL).To(ContainSubstring("finance.yahoo.com"))

		})

	})

	Describe("RefreshSession", func() {

		It("should successfully refresh the HTTP client session", func() {

			mockResponseForCookieSuccess()
			mockResponseForCrumbSuccess()

			outputErr := c.RefreshSession(resty.New(), clientResty)

			Expect(outputErr).ToNot(HaveOccurred())

		})

		Describe("getCookie", func() {

			When("there is an unexpected response when getting a cookie", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseHTTPError("GET", "https://finance.yahoo.com")
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is a client error", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseClientError("GET", "https://finance.yahoo.com")
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is an unexpected response when getting a cookie", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseForCookieUnexpectedResponseError()
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

		})

		Describe("getCookieEU", func() {

			It("should accept the policy and set the cookie", func() {

				var clientRestyAlt = resty.New()
				httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
				defer httpmock.DeactivateAndReset()

				mockResponseForEUConsentRedirect()
				mockResponseForEUConsentRedirectSessionId()
				mockResponseForEUConsentAccept()
				mockResponseForCrumbSuccess()

				outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

				Expect(outputErr).ToNot(HaveOccurred())

			})

			When("there is an unexpected response when getting the session id and CSRF token", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseForEUConsentRedirect()
					mockResponseHTTPError("GET", "https://guce.yahoo.com/consent")
					mockResponseForEUConsentAccept()
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is a client error", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseForEUConsentRedirect()
					mockResponseClientError("GET", "https://guce.yahoo.com/consent")
					mockResponseForEUConsentAccept()
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is no session id in the redirect URL", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseForEUConsentRedirect()

					httpmock.RegisterResponder("GET", "https://guce.yahoo.com/consent", func(request *http.Request) (*http.Response, error) {

						response := httpmock.NewStringResponse(http.StatusFound, "")
						response.Header.Set("Location", "https://consent.yahoo.com/collectConsent?lang=en-US&inline=false")
						response.Request = request

						return response, nil
					})

					httpmock.RegisterResponder("GET", "https://consent.yahoo.com/collectConsent", func(request *http.Request) (*http.Response, error) {

						response := httpmock.NewStringResponse(http.StatusOK, "")

						response.Request = request
						return response, nil
					})

					mockResponseForEUConsentAccept()
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is no CSRF token in the Location header", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(request *http.Request) (*http.Response, error) {

						response := httpmock.NewStringResponse(http.StatusTemporaryRedirect, "")
						response.Header.Set("Location", "https://guce.yahoo.com/consent?brandType=nonEu")
						response.Header.Set("Set-Cookie", "GUCS=TUygQ0Y0; Max-Age=1800; Domain=.yahoo.com; Path=/; Secure")

						response.Request = request
						return response, nil
					})

					mockResponseForEUConsentRedirectSessionId()
					mockResponseForEUConsentAccept()
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is no GUCS cookie set", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(request *http.Request) (*http.Response, error) {

						response := httpmock.NewStringResponse(http.StatusTemporaryRedirect, "")
						response.Header.Set("Location", "https://guce.yahoo.com/consent?brandType=nonEu&gcrumb=HJBaI422")

						response.Request = request
						return response, nil
					})

					mockResponseForEUConsentRedirectSessionId()
					mockResponseForEUConsentAccept()
					mockResponseForCrumbSuccess()

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

		})

		Describe("getCrumb", func() {

			When("there is an unexpected response when getting a crumb", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseForCookieSuccess()
					mockResponseHTTPError("GET", "https://query2.finance.yahoo.com/v1/test/getcrumb")

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

			When("there is a client error", func() {

				It("should return an error", func() {

					var clientRestyAlt = resty.New()
					httpmock.ActivateNonDefault(clientRestyAlt.GetClient())
					defer httpmock.DeactivateAndReset()

					mockResponseForCookieSuccess()
					mockResponseClientError("GET", "https://query2.finance.yahoo.com/v1/test/getcrumb")

					outputErr := c.RefreshSession(resty.New(), clientRestyAlt)

					Expect(outputErr).To(HaveOccurred())

				})

			})

		})

	})

})

func mockResponseForCookieSuccess() {
	httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(request *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, "")
		response.Header.Set("Set-Cookie", "A3=d=AQABBPMJfWQCWPnJSAFIwq1PtsjJQ_yNsJ8FEgEBAQFbfmSGZNxN0iMA_eMAAA&S=AQAAAk_fgKYu72Cro5IHlbBd6yg; Expires=Tue, 4 Jun 2024 04:02:28 GMT; Max-Age=31557600; Domain=.yahoo.com; Path=/; SameSite=None; Secure; HttpOnly")
		return response, nil
	})
}

func mockResponseForCookieUnexpectedResponseError() {
	httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(request *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, "")
		response.Header.Set("Set-Cookie", "")
		return response, nil
	})
}

func mockResponseForCrumbSuccess() {
	httpmock.RegisterResponder("GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", func(request *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, "MrBKM4QQ"), nil
	})
}

func mockResponseHTTPError(method string, url string) {
	httpmock.RegisterResponder(method, url, func(request *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(400, "tests")
		return response, nil
	})
}

func mockResponseClientError(method string, url string) {
	httpmock.RegisterResponder(method, url, func(request *http.Request) (*http.Response, error) {
		return nil, errors.New("test error")
	})
}

func mockResponseForEUConsentRedirect() {
	httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(request *http.Request) (*http.Response, error) {

		response := httpmock.NewStringResponse(http.StatusTemporaryRedirect, "")
		response.Header.Set("Location", "https://guce.yahoo.com/consent?brandType=nonEu&gcrumb=HJBaI422")
		response.Header.Set("Set-Cookie", "GUCS=TUygQ0Y0; Max-Age=1800; Domain=.yahoo.com; Path=/; Secure")

		response.Request = request
		return response, nil
	})
}

func mockResponseForEUConsentRedirectSessionId() {
	httpmock.RegisterResponder("GET", "https://guce.yahoo.com/consent", func(request *http.Request) (*http.Response, error) {

		response := httpmock.NewStringResponse(http.StatusFound, "")
		response.Header.Set("Location", "https://consent.yahoo.com/collectConsent?sessionId=3_cc-session_1_0a0ac970-1b71-4c80-89f2-c2173b355912&lang=en-US&inline=false")
		response.Request = request

		return response, nil
	})

	httpmock.RegisterResponder("GET", "https://consent.yahoo.com/collectConsent", func(request *http.Request) (*http.Response, error) {

		response := httpmock.NewStringResponse(http.StatusOK, "")

		response.Request = request
		return response, nil
	})
}

func mockResponseForEUConsentAccept() {

	httpmock.RegisterResponder("POST", "https://consent.yahoo.com/v2/collectConsent", func(request *http.Request) (*http.Response, error) {

		response := httpmock.NewStringResponse(http.StatusFound, "")
		response.Header.Set("Location", "https://guce.yahoo.com/copyConsent?sessionId=3_cc-session_1_0a0ac970-1b71-4c80-89f2-c2173b355912&lang=da-DK")

		response.Request = request
		return response, nil
	})

	httpmock.RegisterResponder("GET", "https://guce.yahoo.com/copyConsent", func(request *http.Request) (*http.Response, error) {

		response := httpmock.NewStringResponse(http.StatusFound, "")
		response.Header.Set("Location", "https://finance.yahoo.com/")
		response.Header.Set("Set-Cookie", "A3=d=AQABBPMJfWQCWPnJSAFIwq1PtsjJQ_yNsJ8FEgEBAQFbfmSGZNxN0iMA_eMAAA&S=AQAAAk_fgKYu72Cro5IHlbBd6yg; Expires=Tue, 4 Jun 2024 04:02:28 GMT; Max-Age=31557600; Domain=.yahoo.com; Path=/; SameSite=None; Secure; HttpOnly")

		response.Request = request
		return response, nil
	})

}

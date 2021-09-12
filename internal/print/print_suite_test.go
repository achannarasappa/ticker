package print_test

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
)

var client = resty.New()

func mockResponse() {
	response := `{
		"quoteResponse": {
			"result": [
				{
					"quoteType": "EQUITY",
					"currency": "USD",
					"marketState": "CLOSED",
					"shortName": "Alphabet Inc.",
					"preMarketChange": null,
					"preMarketChangePercent": null,
					"regularMarketChange": -59.850098,
					"regularMarketChangePercent": -2.0650284,
					"regularMarketPrice": 2838.42,
					"regularMarketDayHigh": 2920.27,
					"regularMarketDayLow": 2834.83,
					"regularMarketVolume": 1644831,
					"regularMarketPreviousClose": 2898.27,
					"fullExchangeName": "NasdaqGS",
					"regularMarketOpen": 2908.87,
					"fiftyTwoWeekLow": 1406.55,
					"fiftyTwoWeekHigh": 2936.41,
					"marketCap": 1885287088128,
					"exchangeDataDelayedBy": 0,
					"symbol": "GOOG"
				},
				{
					"quoteType": "EQUITY",
					"currency": "USD",
					"marketState": "CLOSED",
					"shortName": "Roblox Corporation",
					"preMarketChange": null,
					"preMarketChangePercent": null,
					"regularMarketChange": 1.5299988,
					"regularMarketChangePercent": 1.7718574,
					"regularMarketPrice": 87.88,
					"regularMarketDayHigh": 90.43,
					"regularMarketDayLow": 84.67,
					"regularMarketVolume": 17465966,
					"regularMarketPreviousClose": 86.35,
					"fullExchangeName": "NYSE",
					"regularMarketOpen": 86.75,
					"fiftyTwoWeekLow": 60.5,
					"fiftyTwoWeekHigh": 103.866,
					"marketCap": 50544357376,
					"exchangeDataDelayedBy": 0,
					"symbol": "RBLX"
				}
			],
			"error": null
		}
	}`
	responseURL := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=GOOG,RBLX"
	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")
		return resp, nil
	})
}

var _ = BeforeSuite(func() {
	httpmock.ActivateNonDefault(client.GetClient())
	mockResponse()
})

var _ = BeforeEach(func() {
	httpmock.Reset()
})

var _ = AfterSuite(func() {
	httpmock.DeactivateAndReset()
})

func TestPrint(t *testing.T) {
	format.TruncatedDiff = false
	RegisterFailHandler(Fail)
	RunSpecs(t, "Print Suite")
}

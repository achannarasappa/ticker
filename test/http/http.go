package http

import (
	"bytes"
	"errors"
	"html/template"
	"net/http"

	"github.com/jarcoal/httpmock"
)

// ResponseParameters represents response values for a templated HTTP API response
type ResponseParameters struct {
	Symbol   string
	Currency string
	Price    float64
}

// MockResponse registers a mock responder for price quotes
func MockResponse(responseParameters ResponseParameters) {
	var responseBytes bytes.Buffer
	responseTemplate := `{
		"quoteResponse": {
			"result": [
				{
					"regularMarketPrice": {
						"raw": {{.Price}},
						"fmt": "{{.Price}}"
					},
					"currency": "{{.Currency}}",
					"symbol": "{{.Symbol}}"
				}
			],
			"error": null
		}
	}`
	t, _ := template.New("response").Parse(responseTemplate)
	//nolint:errcheck
	t.Execute(&responseBytes, responseParameters)
	responseUrl := `=~\/finance\/quote.*fields\=regularMarketPrice%2Ccurrency\&symbols\=` + responseParameters.Symbol //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, responseBytes.String())
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

// MockResponseCurrency registers a mock responder for currency rates
func MockResponseCurrency() {
	response := `{
		"quoteResponse": {
			"result": [
				{
					"regularMarketPrice": {"raw": 123.45, "fmt": "123.45"},
					"currency": "USD",
					"symbol": "NET"
				}
			],
			"error": null
		}
	}`
	responseUrl := `=~.*\/finance\/quote.*` //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

func MockResponseCurrencyError() {
	responseUrl := `=~.*\/finance\/quote.*` //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		return &http.Response{}, errors.New("error getting currencies") //nolint:goerr113
	})
}

func MockResponseYahooQuotes() {
	response := `{
		"quoteResponse": {
			"result": [{
					"quoteType": "EQUITY",
					"currency": "USD",
					"marketState": "CLOSED",
					"shortName": "Alphabet Inc.",
					"regularMarketChange": {"raw": -59.850098, "fmt": "-59.850098"},
					"regularMarketChangePercent": {"raw": -2.0650284, "fmt": "-2.0650284"},
					"regularMarketPrice": {"raw": 2838.42, "fmt": "2838.42"},
					"regularMarketDayHigh": {"raw": 2920.27, "fmt": "2920.27"},
					"regularMarketDayLow": {"raw": 2834.83, "fmt": "2834.83"},
					"regularMarketVolume": {"raw": 1644831, "fmt": "1644831"},
					"regularMarketPreviousClose": {"raw": 2898.27, "fmt": "2898.27"},
					"fullExchangeName": "NasdaqGS",
					"regularMarketOpen": {"raw": 2908.87, "fmt": "2908.87"},
					"fiftyTwoWeekLow": {"raw": 1406.55, "fmt": "1406.55"},
					"fiftyTwoWeekHigh": {"raw": 2936.41, "fmt": "2936.41"},
					"marketCap": {"raw": 1885287088128, "fmt": "1.8bn"},
					"exchangeDataDelayedBy": 0,
					"symbol": "GOOG"
				},
				{
					"quoteType": "EQUITY",
					"currency": "USD",
					"marketState": "CLOSED",
					"shortName": "Roblox Corporation",
					"regularMarketChange": {"raw": 1.5299988, "fmt": "1.5299988"},
					"regularMarketChangePercent": {"raw": 1.7718574, "fmt": "1.7718574"},
					"regularMarketPrice": {"raw": 87.88, "fmt": "87.88"},
					"regularMarketDayHigh": {"raw": 90.43, "fmt": "90.43"},
					"regularMarketDayLow": {"raw": 84.67, "fmt": "84.67"},
					"regularMarketVolume": {"raw": 17465966, "fmt": "17465966"},
					"regularMarketPreviousClose": {"raw": 86.35, "fmt": "86.35"},
					"fullExchangeName": "NYSE",
					"regularMarketOpen": {"raw": 86.75, "fmt": "86.75"},
					"fiftyTwoWeekLow": {"raw": 60.5, "fmt": "60.5"},
					"fiftyTwoWeekHigh": {"raw": 103.866, "fmt": "103.866"},
					"marketCap": {"raw": 50544357376, "fmt": "5.0bn"},
					"exchangeDataDelayedBy": 0,
					"symbol": "RBLX"
				}
			],
			"error": null
		}
	}`
	responseUrl := `=~.*\/finance\/quote.*symbols\=GOOG.*RBLX.*` //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

func MockTickerSymbols() {
	responseFixture := `"ADA.X","cardano","cg"
"ALGO.X","algorand","cg"
"BTC.X","bitcoin","cg"
"ETH.X","ethereum","cg"
"DOGE.X","dogecoin","cg"
"DOT.X","polkadot","cg"
"SOL.X","solana","cg"
"USDC.X","usd-coin","cg"
"XRP.X","ripple","cg"
`
	responseUrl := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv" //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, responseFixture)
		resp.Header.Set("Content-Type", "text/plain; charset=utf-8")

		return resp, nil
	})
}

func MockTickerSymbolsError() {
	responseUrl := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv" //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		return &http.Response{}, errors.New("error getting ticker symbols") //nolint:goerr113
	})
}

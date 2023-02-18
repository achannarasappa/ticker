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
					"regularMarketPrice": {{.Price}},
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
	responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=" + responseParameters.Symbol //nolint:golint,stylecheck,revive
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
					"regularMarketPrice": 123.45,
					"currency": "USD",
					"symbol": "NET"
				}
			],
			"error": null
		}
	}`
	responseUrl := `https://query1.finance.yahoo.com/v7/finance/quote` //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

func MockResponseCurrencyError() {
	responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote" //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
		return &http.Response{}, errors.New("error getting currencies") //nolint:goerr113
	})
}

func MockResponseYahooQuotes() {
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
	responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?fields=shortName,regularMarketChange,regularMarketChangePercent,regularMarketPrice,regularMarketPreviousClose,regularMarketOpen,regularMarketDayRange,regularMarketDayHigh,regularMarketDayLow,regularMarketVolume,postMarketChange,postMarketChangePercent,postMarketPrice,preMarketChange,preMarketChangePercent,preMarketPrice,fiftyTwoWeekHigh,fiftyTwoWeekLow,marketCap&region=US&lang=en-US&symbols=GOOG,RBLX" //nolint:golint,stylecheck,revive
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

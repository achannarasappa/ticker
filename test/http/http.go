package http

import (
	"bytes"
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
	responseURL := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=" + responseParameters.Symbol
	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
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
	responseURL := `https://query1.finance.yahoo.com/v7/finance/quote`
	httpmock.RegisterResponder("GET", responseURL, func(req *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")
		return resp, nil
	})
}

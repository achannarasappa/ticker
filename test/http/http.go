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
	httpmock.RegisterResponder("GET", responseUrl, func(_ *http.Request) (*http.Response, error) {
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
	httpmock.RegisterResponder("GET", responseUrl, func(_ *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

func MockResponseCurrencyError() {
	responseUrl := `=~.*\/finance\/quote.*` //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseUrl, func(_ *http.Request) (*http.Response, error) {

		return httpmock.NewStringResponse(http.StatusInternalServerError, "server error"), errors.New("error getting currencies") //nolint:goerr113
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
	httpmock.RegisterResponder("GET", responseUrl, func(_ *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, response)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

func MockResponseCoingeckoQuotes() {

	responseFixture := `[
				{
						"ath": 69045,
						"ath_change_percentage": -43.4461,
						"ath_date": "2021-11-10T14:24:11.849Z",
						"atl": 67.81,
						"atl_change_percentage": 57484.55501,
						"atl_date": "2013-07-06T00:00:00.000Z",
						"circulating_supply": 18964093.0,
						"current_price": 39045,
						"fully_diluted_valuation": 819997729028,
						"high_24h": 40090,
						"id": "bitcoin",
						"image": "https://assets.coingecko.com/coins/images/1/large/bitcoin.png?1547033579",
						"last_updated": "2022-02-21T01:24:23.221Z",
						"low_24h": 38195,
						"market_cap": 740500628242,
						"market_cap_change_24h": -17241577635.956177,
						"market_cap_change_percentage_24h": -2.27539,
						"market_cap_rank": 1,
						"max_supply": 21000000.0,
						"name": "Bitcoin",
						"price_change_24h": -978.048909591315,
						"price_change_percentage_24h": -2.44373,
						"roi": null,
						"symbol": "btc",
						"total_supply": 21000000.0,
						"total_volume": 16659222262
				}
		]`
	responseURL := "https://api.coingecko.com/api/v3/coins/markets?vs_currency=usd&ids=bitcoin&order=market_cap_desc&per_page=250&page=1&sparkline=false"
	httpmock.RegisterResponder("GET", responseURL, func(_ *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, responseFixture)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})
}

func MockResponseCoincapQuotes() {

	responseFixture := `{
		"data": [
			{
				"id": "elrond",
				"rank": "1",
				"symbol": "EGLD",
				"name": "MultiversX",
				"supply": "19685775.0000000000000000",
				"maxSupply": "21000000.0000000000000000",
				"marketCapUsd": "1248489381324.9592799671502700",
				"volumeUsd24Hr": "7744198446.5431034815177485",
				"priceUsd": "63420.8905326287270868",
				"changePercent24Hr": "1.3622077494913284",
				"vwap24Hr": "62988.1090433238215198",
				"explorer": "https://blockchain.info/"
			}
		],
		"timestamp": 1714453771801
	}`
	responseURL := `=~\/v2\/assets.*ids\=elrond.*`
	httpmock.RegisterResponder("GET", responseURL, func(_ *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, responseFixture)
		resp.Header.Set("Content-Type", "application/json")

		return resp, nil
	})

}

func MockResponseCoinbaseQuotes() {
	responseFixture := `{
		"products": [
			{
				"product_id": "ADA-31JAN25-CDE",
				"price": "97345",
				"price_percentage_change_24h": "-3.14412218297597",
				"volume_24h": "93744",
				"base_name": "",
				"status": "",
				"product_type": "FUTURE",
				"quote_currency_id": "USD",
				"fcm_trading_session_details": {
					"is_session_open": true
				},
				"base_display_symbol": "",
				"product_venue": "FCM",
				"future_product_details": {
					"venue": "cde",
					"contract_code": "ADA",
					"contract_expiry": "2025-01-31T16:00:00Z",
					"contract_root_unit": "ADA",
					"group_description": "Cardano Futures",
					"contract_expiry_timezone": "Europe/London"
				}
			},
			{
				"base_display_symbol": "ADA",
				"product_type": "SPOT",
				"product_id": "ADA-USD",
				"base_name": "Cardano",
				"price": "50000.00",
				"price_percentage_change_24h": "2.0408163265306123",
				"volume_24h": "1500.50",
				"display_name": "Cardano",
				"status": "",
				"quote_currency_id": "USD",
				"product_venue": "CBE"
			}
		]
	}`
	responseURL := `=~\/api\/v3\/brokerage\/market\/products.*`
	httpmock.RegisterResponder("GET", responseURL, func(_ *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, responseFixture)
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
	responseURL := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv" //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseURL, func(_ *http.Request) (*http.Response, error) {
		resp := httpmock.NewStringResponse(200, responseFixture)
		resp.Header.Set("Content-Type", "text/plain; charset=utf-8")

		return resp, nil
	})
}

func MockTickerSymbolsError() {
	responseURL := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv" //nolint:golint,stylecheck,revive
	httpmock.RegisterResponder("GET", responseURL, func(_ *http.Request) (*http.Response, error) {

		return httpmock.NewStringResponse(http.StatusInternalServerError, "server error"), errors.New("error getting ticker symbols") //nolint:goerr113
	})
}

// Mocks for Yahoo session refresh
func MockResponseForRefreshSessionSuccess() {
	httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(_ *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, "")
		response.Header.Set("Set-Cookie", "A3=d=AQABBPMJfWQCWPnJSAFIwq1PtsjJQ_yNsJ8FEgEBAQFbfmSGZNxN0iMA_eMAAA&S=AQAAAk_fgKYu72Cro5IHlbBd6yg; Expires=Tue, 4 Jun 2024 04:02:28 GMT; Max-Age=31557600; Domain=.yahoo.com; Path=/; SameSite=None; Secure; HttpOnly")

		return response, nil
	})

	httpmock.RegisterResponder("GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", func(_ *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusOK, "MrBKM4QQ"), nil
	})
}

func MockResponseForRefreshSessionError() {
	httpmock.RegisterResponder("GET", "https://finance.yahoo.com/", func(_ *http.Request) (*http.Response, error) {
		response := httpmock.NewStringResponse(http.StatusOK, "")
		response.Header.Set("Set-Cookie", "A3=d=AQABBPMJfWQCWPnJSAFIwq1PtsjJQ_yNsJ8FEgEBAQFbfmSGZNxN0iMA_eMAAA&S=AQAAAk_fgKYu72Cro5IHlbBd6yg; Expires=Tue, 4 Jun 2024 04:02:28 GMT; Max-Age=31557600; Domain=.yahoo.com; Path=/; SameSite=None; Secure; HttpOnly")

		return response, nil
	})

	httpmock.RegisterResponder("GET", "https://query2.finance.yahoo.com/v1/test/getcrumb", func(_ *http.Request) (*http.Response, error) {
		return httpmock.NewStringResponse(http.StatusBadRequest, "tests"), nil
	})
}

package http

import (
	"errors"
	"net/http"

	"github.com/jarcoal/httpmock"
)

// ResponseParameters represents response values for a templated HTTP API response
type ResponseParameters struct {
	Symbol   string
	Currency string
	Price    float64
}

func MockTickerSymbols() {
	responseFixture := `"ADA.X","cardano","cb"
"ALGO.X","algorand","cb"
"BTC.X","bitcoin","cb"
"ETH.X","ethereum","cb"
"DOGE.X","dogecoin","cb"
"DOT.X","polkadot","cb"
"SOL.X","solana","cb"
"USDC.X","usd-coin","cb"
"XRP.X","ripple","cb"
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

package coinbase

import (
	"net/url"

	"github.com/go-resty/resty/v2"
)

// ResponseQuote represents a quote of a single product from the Coinbase API
type ResponseQuote struct {
	Symbol                   string `json:"base_display_symbol"`
	ProductID                string `json:"product_id"`
	ShortName                string `json:"base_name"`
	Price                    string `json:"price"`
	PriceChange24H           string `json:"price_percentage_change_24h"`
	Volume24H                string `json:"volume_24h"`
	DisplayName              string `json:"display_name"`
	MarketState              string `json:"status"`
	Currency                 string `json:"quote_currency_id"`
	ExchangeName             string `json:"product_venue"`
	FcmTradingSessionDetails struct {
		IsSessionOpen bool `json:"is_session_open"`
	} `json:"fcm_trading_session_details"`
	FutureProductDetails struct {
		ContractDisplayName string `json:"contract_display_name"`
		GroupDescription    string `json:"group_description"`
		ContractRootUnit    string `json:"contract_root_unit"`
		ExpirationDate      string `json:"contract_expiry"`
		ExpirationTimezone  string `json:"expiration_timezone"`
	} `json:"future_product_details"`
	ProductType string `json:"product_type"`
}

// Response represents the container object from the API response
type Response struct {
	Products []ResponseQuote `json:"products"`
}

// GetUnderlyingAssetSymbols retrieves the underlying asset symbol for Coinbase futures contracts
// For example, BIT-29NOV24-CDE would return BTC as the underlying asset
func GetUnderlyingAssetSymbols(client resty.Client, symbols []string) ([]string, error) {
	underlyingSymbols := make([]string, 0)

	res, _ := client.R().
		SetResult(Response{}).
		SetQueryParamsFromValues(url.Values{"product_ids": symbols}).
		Get("https://api.coinbase.com/api/v3/brokerage/market/products")

	quotes := res.Result().(*Response).Products //nolint:forcetypeassert

	for _, quote := range quotes {

		if quote.ProductType == "FUTURE" {
			quoteUnderlyingSymbol := quote.FutureProductDetails.ContractRootUnit + "-USD"
			underlyingSymbols = append(underlyingSymbols, quoteUnderlyingSymbol)
		}

	}

	return underlyingSymbols, nil
}

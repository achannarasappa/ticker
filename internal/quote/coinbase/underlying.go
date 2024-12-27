package coinbase

import (
	"net/url"

	"github.com/go-resty/resty/v2"
)

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

package yahoo

import (
	"strings"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

func getCurrencyPair(pair string) (string, string) {
	return pair[:3], pair[3:6]
}

func transformResponseCurrency(responseQuote ResponseQuote) c.CurrencyRate {

	fromCurrency, toCurrency := getCurrencyPair(responseQuote.Symbol)

	return c.CurrencyRate{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         responseQuote.RegularMarketPrice.Raw,
	}

}

func transformResponseCurrencies(responseQuotes []ResponseQuote) c.CurrencyRates {

	currencyRates := c.CurrencyRates{}

	for _, responseQuote := range responseQuotes {
		currencyRate := transformResponseCurrency(responseQuote)
		currencyRates[currencyRate.FromCurrency] = currencyRate
	}

	return currencyRates

}

func getCurrencyRatesFromCurrencyPairSymbols(client resty.Client, currencyPairSymbols []string) (c.CurrencyRates, error) {

	symbolsString := strings.Join(currencyPairSymbols, ",")

	res, err := client.R().
		SetResult(Response{}).
		SetQueryParam("fields", "regularMarketPrice,currency").
		SetQueryParam("symbols", symbolsString).
		Get("/v7/finance/quote")

	if err != nil {
		return c.CurrencyRates{}, err
	}

	return transformResponseCurrencies((res.Result().(*Response)).QuoteResponse.Quotes), nil //nolint:forcetypeassert
}

func transformResponseCurrencyPairs(responseQuotes []ResponseQuote, targetCurrency string) []string {

	targetCurrencyPair := targetCurrency + targetCurrency + "=X"

	keys := make(map[string]bool)
	currencyPairSymbols := make([]string, 0)

	for _, responseQuote := range responseQuotes {
		pair := responseQuote.Currency + targetCurrency + "=X"
		if _, exists := keys[pair]; !exists && pair != targetCurrencyPair && pair != targetCurrency+"=X" {
			keys[pair] = true
			currencyPairSymbols = append(currencyPairSymbols, pair)
		}
	}

	return currencyPairSymbols

}

func getCurrencyPairSymbols(client resty.Client, symbols []string, targetCurrency string) ([]string, error) {

	symbolsString := strings.Join(symbols, ",")

	res, err := client.R().
		SetResult(Response{}).
		SetQueryParam("fields", "regularMarketPrice,currency").
		SetQueryParam("symbols", symbolsString).
		Get("/v7/finance/quote")

	if err != nil {
		return []string{}, err
	}

	return transformResponseCurrencyPairs((res.Result().(*Response)).QuoteResponse.Quotes, targetCurrency), nil //nolint:forcetypeassert
}

// GetCurrencyRates retrieves the currency rates to convert from each currency for the given symbols to the target currency
func GetCurrencyRates(client resty.Client, symbols []string, targetCurrency string) (c.CurrencyRates, error) {

	if targetCurrency == "" {
		targetCurrency = "USD"
	}

	currencyPairSymbols, err := getCurrencyPairSymbols(client, symbols, targetCurrency)

	if err != nil {
		return c.CurrencyRates{}, err
	}

	if len(currencyPairSymbols) == 0 {
		return c.CurrencyRates{}, nil
	}

	currencyRates, err := getCurrencyRatesFromCurrencyPairSymbols(client, currencyPairSymbols)

	if err != nil {
		return c.CurrencyRates{}, err
	}

	return currencyRates, nil
}

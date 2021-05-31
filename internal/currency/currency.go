package currency

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/go-resty/resty/v2"
)

type ResponseQuote struct {
	Symbol             string  `json:"symbol"`
	Currency           string  `json:"currency"`
	RegularMarketPrice float64 `json:"regularMarketPrice"`
}

type Response struct {
	QuoteResponse struct {
		Quotes []ResponseQuote `json:"result"`
		Error  interface{}     `json:"error"`
	} `json:"quoteResponse"`
}

type CurrencyRateByUse struct {
	ToCurrencyCode string
	QuotePrice     float64
	PositionValue  float64
	PositionCost   float64
	SummaryValue   float64
	SummaryCost    float64
}

func getCurrencyPair(pair string) (string, string) {
	return pair[:3], pair[3:6]
}

func transformResponseCurrency(responseQuote ResponseQuote) c.CurrencyRate {

	fromCurrency, toCurrency := getCurrencyPair(responseQuote.Symbol)

	return c.CurrencyRate{
		FromCurrency: fromCurrency,
		ToCurrency:   toCurrency,
		Rate:         responseQuote.RegularMarketPrice,
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
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=%s", symbolsString)

	res, err := client.R().
		SetResult(Response{}).
		Get(url)

	if err != nil {
		return c.CurrencyRates{}, err
	}

	return transformResponseCurrencies((res.Result().(*Response)).QuoteResponse.Quotes), nil
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
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=%s", symbolsString)
	res, err := client.R().
		SetResult(Response{}).
		Get(url)

	if err != nil {
		return []string{}, err
	}

	return transformResponseCurrencyPairs((res.Result().(*Response)).QuoteResponse.Quotes, targetCurrency), nil
}

func GetCurrencyRates(client resty.Client, symbols []string, targetCurrency string) (c.CurrencyRates, error) {

	if targetCurrency == "" {
		targetCurrency = "USD"
	}

	currencyPairSymbols, err := getCurrencyPairSymbols(client, symbols, targetCurrency)

	if err != nil {
		return c.CurrencyRates{}, err
	}

	if len(currencyPairSymbols) <= 0 {
		return c.CurrencyRates{}, nil
	}

	currencyRates, err := getCurrencyRatesFromCurrencyPairSymbols(client, currencyPairSymbols)

	if err != nil {
		return c.CurrencyRates{}, err
	}

	return currencyRates, nil
}

func GetCurrencyRateFromContext(ctx c.Context, fromCurrency string) CurrencyRateByUse {
	if currencyRate, ok := ctx.Reference.CurrencyRates[fromCurrency]; ok {

		// Convert only the summary currency to the configured currency
		if ctx.Config.Currency != "" && ctx.Config.CurrencyConvertSummaryOnly {
			return CurrencyRateByUse{
				ToCurrencyCode: fromCurrency,
				QuotePrice:     1.0,
				PositionValue:  1.0,
				PositionCost:   1.0,
				SummaryValue:   currencyRate.Rate,
				SummaryCost:    currencyRate.Rate,
			}
		}

		// Convert all quotes and positions to target currency and implicitly convert summary currency (i.e. no conversion since underlying values are already converted)
		if ctx.Config.Currency != "" {
			return CurrencyRateByUse{
				ToCurrencyCode: currencyRate.ToCurrency,
				QuotePrice:     currencyRate.Rate,
				PositionValue:  currencyRate.Rate,
				PositionCost:   currencyRate.Rate,
				SummaryValue:   1.0,
				SummaryCost:    1.0,
			}
		}

		// Convert only the summary currency to the default currency (USD) when currency conversion is not enabled
		return CurrencyRateByUse{
			ToCurrencyCode: currencyRate.ToCurrency,
			QuotePrice:     1.0,
			PositionValue:  1.0,
			PositionCost:   1.0,
			SummaryValue:   currencyRate.Rate,
			SummaryCost:    currencyRate.Rate,
		}

	}
	return CurrencyRateByUse{
		ToCurrencyCode: fromCurrency,
		QuotePrice:     1.0,
		PositionValue:  1.0,
		PositionCost:   1.0,
		SummaryValue:   1.0,
		SummaryCost:    1.0,
	}
}

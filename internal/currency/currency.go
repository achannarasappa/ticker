package currency

import (
	"fmt"
	"strings"

	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/go-resty/resty/v2"
	. "github.com/novalagung/gubrak/v2"
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

	currencyRates := From(responseQuotes).Reduce(func(acc c.CurrencyRates, responseQuote ResponseQuote) c.CurrencyRates {
		currencyRate := transformResponseCurrency(responseQuote)
		acc[currencyRate.FromCurrency] = currencyRate
		return acc
	}, c.CurrencyRates{}).Result()

	return (currencyRates).(c.CurrencyRates)

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

	currencyPairSymbols := From(responseQuotes).
		Map(func(v ResponseQuote) string {
			return v.Currency + targetCurrency + "=X"
		}).
		Uniq().
		Reject(func(v string) bool {
			return v == targetCurrencyPair
		}).
		Result()

	return (currencyPairSymbols).([]string)

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

func GetCurrencyRateFromContext(ctx c.Context, fromCurrency string) (float64, float64, string) {
	if currencyRate, ok := ctx.Reference.CurrencyRates[fromCurrency]; ok {
		if ctx.Config.Currency != "" {
			return currencyRate.Rate, 1.0, currencyRate.ToCurrency
		}

		return 1.0, currencyRate.Rate, currencyRate.ToCurrency

	}
	return 1.0, 1.0, fromCurrency
}

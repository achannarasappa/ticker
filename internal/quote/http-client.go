package quote

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Quote struct {
	Currency                   string  `json:"currency"`
	ShortName                  string  `json:"shortName"`
	Symbol                     string  `json:"symbol"`
	MarketState                string  `json:"marketState"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	PostMarketChange           float64 `json:"postMarketChange"`
	PostMarketChangePercent    float64 `json:"postMarketChangePercent"`
	PostMarketPrice            float64 `json:"postMarketPrice"`
}

type Response struct {
	QuoteResponse struct {
		Quotes []Quote     `json:"result"`
		Error  interface{} `json:"error"`
	} `json:"quoteResponse"`
}

func GetQuotes(symbols []string) []Quote {
	symbolsString := strings.Join(symbols, ",")
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&symbols=%s", symbolsString)
	response, _ := resty.New().R().
		SetResult(&Response{}).
		Get(url)

	return (response.Result().(*Response)).QuoteResponse.Quotes
}

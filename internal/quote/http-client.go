package quote

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Quote struct {
	Currency                   string  `json:"currency"`
	MarketCap                  int64   `json:"marketCap"`
	ShortName                  string  `json:"shortName"`
	MarketState                string  `json:"marketState"`
	RegularMarketChange        float64 `json:"regularMarketChange"`
	RegularMarketChangePercent float64 `json:"regularMarketChangePercent"`
	RegularMarketPrice         float64 `json:"regularMarketPrice"`
	RegularMarketDayHigh       float64 `json:"regularMarketDayHigh"`
	RegularMarketDayLow        float64 `json:"regularMarketDayLow"`
	RegularMarketOpen          float64 `json:"regularMarketOpen"`
	Symbol                     string  `json:"symbol"`
	RegularMarketVolume        int     `json:"regularMarketVolume"`
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

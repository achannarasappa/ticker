package unary

// Response represents the container object from the API response
type Response struct {
	QuoteResponse ResponseQuoteResponse `json:"quoteResponse"`
}

type ResponseQuoteResponse struct {
	Quotes []ResponseQuote `json:"result"`
	Error  interface{}     `json:"error"`
}

// ResponseQuote represents a quote of a single security from the API response
type ResponseQuote struct {
	ShortName                  string              `json:"shortName"`
	Symbol                     string              `json:"symbol"`
	MarketState                string              `json:"marketState"`
	Currency                   string              `json:"currency"`
	ExchangeName               string              `json:"fullExchangeName"`
	ExchangeDelay              float64             `json:"exchangeDataDelayedBy"`
	RegularMarketChange        ResponseFieldFloat  `json:"regularMarketChange"`
	RegularMarketChangePercent ResponseFieldFloat  `json:"regularMarketChangePercent"`
	RegularMarketPrice         ResponseFieldFloat  `json:"regularMarketPrice"`
	RegularMarketPreviousClose ResponseFieldFloat  `json:"regularMarketPreviousClose"`
	RegularMarketOpen          ResponseFieldFloat  `json:"regularMarketOpen"`
	RegularMarketDayRange      ResponseFieldString `json:"regularMarketDayRange"`
	RegularMarketDayHigh       ResponseFieldFloat  `json:"regularMarketDayHigh"`
	RegularMarketDayLow        ResponseFieldFloat  `json:"regularMarketDayLow"`
	RegularMarketVolume        ResponseFieldFloat  `json:"regularMarketVolume"`
	PostMarketChange           ResponseFieldFloat  `json:"postMarketChange"`
	PostMarketChangePercent    ResponseFieldFloat  `json:"postMarketChangePercent"`
	PostMarketPrice            ResponseFieldFloat  `json:"postMarketPrice"`
	PreMarketChange            ResponseFieldFloat  `json:"preMarketChange"`
	PreMarketChangePercent     ResponseFieldFloat  `json:"preMarketChangePercent"`
	PreMarketPrice             ResponseFieldFloat  `json:"preMarketPrice"`
	FiftyTwoWeekHigh           ResponseFieldFloat  `json:"fiftyTwoWeekHigh"`
	FiftyTwoWeekLow            ResponseFieldFloat  `json:"fiftyTwoWeekLow"`
	QuoteType                  string              `json:"quoteType"`
	MarketCap                  ResponseFieldFloat  `json:"marketCap"`
}

type ResponseFieldFloat struct {
	Raw float64 `json:"raw"`
	Fmt string  `json:"fmt"`
}

type ResponseFieldString struct {
	Raw string `json:"raw"`
	Fmt string `json:"fmt"`
}

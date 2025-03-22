package unary_test

import (
	"github.com/achannarasappa/ticker/v4/internal/monitor/yahoo/unary"
)

var (
	responseQuote1Fixture = unary.Response{
		QuoteResponse: unary.ResponseQuoteResponse{
			Quotes: []unary.ResponseQuote{
				{
					MarketState:                "REGULAR",
					ShortName:                  "Cloudflare, Inc.",
					PreMarketChange:            unary.ResponseFieldFloat{Raw: 1.0399933, Fmt: "1.0399933"},
					PreMarketChangePercent:     unary.ResponseFieldFloat{Raw: 1.2238094, Fmt: "1.2238094"},
					PreMarketPrice:             unary.ResponseFieldFloat{Raw: 86.03, Fmt: "86.03"},
					RegularMarketChange:        unary.ResponseFieldFloat{Raw: 3.0800018, Fmt: "3.0800018"},
					RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: 3.7606857, Fmt: "3.7606857"},
					RegularMarketPrice:         unary.ResponseFieldFloat{Raw: 84.98, Fmt: "84.98"},
					RegularMarketPreviousClose: unary.ResponseFieldFloat{Raw: 84.00, Fmt: "84.00"},
					RegularMarketOpen:          unary.ResponseFieldFloat{Raw: 85.22, Fmt: "85.22"},
					RegularMarketDayHigh:       unary.ResponseFieldFloat{Raw: 90.00, Fmt: "90.00"},
					RegularMarketDayLow:        unary.ResponseFieldFloat{Raw: 80.00, Fmt: "80.00"},
					PostMarketChange:           unary.ResponseFieldFloat{Raw: 1.37627, Fmt: "1.37627"},
					PostMarketChangePercent:    unary.ResponseFieldFloat{Raw: 1.35735, Fmt: "1.35735"},
					PostMarketPrice:            unary.ResponseFieldFloat{Raw: 86.56, Fmt: "86.56"},
					Symbol:                     "NET",
				},
			},
			Error: nil,
		},
	}
)

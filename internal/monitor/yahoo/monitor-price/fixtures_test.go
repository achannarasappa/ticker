package monitorPriceYahoo_test

import (
	"github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"
)

var (
	responseQuote1Fixture = unary.Response{
		QuoteResponse: unary.ResponseQuoteResponse{
			Quotes: []unary.ResponseQuote{
				quoteCloudflareFixture,
				quoteGoogleFixture,
			},
			Error: nil,
		},
	}
	currencyResponseFixture = unary.Response{
		QuoteResponse: unary.ResponseQuoteResponse{
			Quotes: []unary.ResponseQuote{
				{
					RegularMarketPrice: unary.ResponseFieldFloat{Raw: 1.25, Fmt: "1.25"},
					Currency:           "EUR",
					Symbol:             "EURUSD=X",
				},
				{
					RegularMarketPrice: unary.ResponseFieldFloat{Raw: 0.92, Fmt: "0.92"},
					Currency:           "USD",
					Symbol:             "USDEUR=X",
				},
			},
			Error: nil,
		},
	}
	quoteCloudflareFixture = unary.ResponseQuote{
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
	}
	quoteGoogleFixture = unary.ResponseQuote{
		MarketState:                "REGULAR",
		ShortName:                  "Google Inc.",
		PreMarketChange:            unary.ResponseFieldFloat{Raw: 1.0399933, Fmt: "1.0399933"},
		PreMarketChangePercent:     unary.ResponseFieldFloat{Raw: 1.2238094, Fmt: "1.2238094"},
		PreMarketPrice:             unary.ResponseFieldFloat{Raw: 166.03, Fmt: "166.03"},
		RegularMarketChange:        unary.ResponseFieldFloat{Raw: 3.0800018, Fmt: "3.0800018"},
		RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: 3.7606857, Fmt: "3.7606857"},
		RegularMarketPrice:         unary.ResponseFieldFloat{Raw: 166.25, Fmt: "166.25"},
		RegularMarketPreviousClose: unary.ResponseFieldFloat{Raw: 165.00, Fmt: "165.00"},
		RegularMarketOpen:          unary.ResponseFieldFloat{Raw: 165.00, Fmt: "165.00"},
		RegularMarketDayHigh:       unary.ResponseFieldFloat{Raw: 167.00, Fmt: "167.00"},
		RegularMarketDayLow:        unary.ResponseFieldFloat{Raw: 164.00, Fmt: "164.00"},
		PostMarketChange:           unary.ResponseFieldFloat{Raw: 1.37627, Fmt: "1.37627"},
		PostMarketChangePercent:    unary.ResponseFieldFloat{Raw: 1.35735, Fmt: "1.35735"},
		PostMarketPrice:            unary.ResponseFieldFloat{Raw: 167.62, Fmt: "167.62"},
		Symbol:                     "GOOG",
	}
	quoteMetaFixture = unary.ResponseQuote{
		MarketState:                "REGULAR",
		ShortName:                  "Meta Platforms Inc.",
		PreMarketChange:            unary.ResponseFieldFloat{Raw: 2, Fmt: "2"},
		PreMarketChangePercent:     unary.ResponseFieldFloat{Raw: 0.6666667, Fmt: "0.6666667"},
		PreMarketPrice:             unary.ResponseFieldFloat{Raw: 300.00, Fmt: "300.00"},
		RegularMarketChange:        unary.ResponseFieldFloat{Raw: 3.00, Fmt: "3.00"},
		RegularMarketChangePercent: unary.ResponseFieldFloat{Raw: 1.00, Fmt: "1.00"},
		RegularMarketPrice:         unary.ResponseFieldFloat{Raw: 303.00, Fmt: "303.00"},
		RegularMarketPreviousClose: unary.ResponseFieldFloat{Raw: 300.00, Fmt: "300.00"},
		RegularMarketOpen:          unary.ResponseFieldFloat{Raw: 300.00, Fmt: "300.00"},
		RegularMarketDayHigh:       unary.ResponseFieldFloat{Raw: 305.00, Fmt: "305.00"},
		RegularMarketDayLow:        unary.ResponseFieldFloat{Raw: 295.00, Fmt: "295.00"},
		PostMarketChange:           unary.ResponseFieldFloat{Raw: 1.00, Fmt: "1.00"},
		PostMarketChangePercent:    unary.ResponseFieldFloat{Raw: 0.3333333, Fmt: "0.3333333"},
		PostMarketPrice:            unary.ResponseFieldFloat{Raw: 304.37, Fmt: "304.37"},
		Symbol:                     "FB",
	}
)

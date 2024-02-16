package asset_test

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
)

var fixtureAssetGroupQuote = c.AssetGroupQuote{
	AssetGroup: c.AssetGroup{
		ConfigAssetGroup: c.ConfigAssetGroup{
			Name: "default",
			Watchlist: []string{
				"TWKS",
				"MSFT",
				"SOL1-USD",
			},
		},
	},
	AssetQuotes: []c.AssetQuote{
		{
			Name:     "ThoughtWorks",
			Symbol:   "TWKS",
			Class:    c.AssetClassStock,
			Currency: c.Currency{FromCurrencyCode: "USD"},
			QuotePrice: c.QuotePrice{
				Price:          110.0,
				PricePrevClose: 100.0,
				PriceOpen:      100.0,
				PriceDayHigh:   110.0,
				PriceDayLow:    90.0,
				Change:         10.0,
				ChangePercent:  10.0,
			},
			QuoteExtended: c.QuoteExtended{
				FiftyTwoWeekHigh: 150,
				FiftyTwoWeekLow:  50,
				MarketCap:        1000000,
			},
		},
		{
			Name:     "Microsoft Inc",
			Symbol:   "MSFT",
			Class:    c.AssetClassStock,
			Currency: c.Currency{FromCurrencyCode: "USD"},
			QuotePrice: c.QuotePrice{
				Price:          220.0,
				PricePrevClose: 200.0,
				PriceOpen:      200.0,
				PriceDayHigh:   220.0,
				PriceDayLow:    180.0,
				Change:         20.0,
				ChangePercent:  10.0,
			},
		},
		{
			Name:     "Solana USD",
			Symbol:   "SOL1-USD",
			Class:    c.AssetClassCryptocurrency,
			Currency: c.Currency{FromCurrencyCode: "USD"},
			QuotePrice: c.QuotePrice{
				Price:          11.0,
				PricePrevClose: 10.0,
				PriceOpen:      10.0,
				PriceDayHigh:   11.0,
				PriceDayLow:    9.0,
				Change:         1.0,
				ChangePercent:  10.0,
			},
			Meta: c.Meta{
				IsVariablePrecision: true,
			},
		},
	},
}
var fixtureAssets = []c.Asset{
	{
		Name:     "ThoughtWorks",
		Symbol:   "TWKS",
		Class:    c.AssetClassStock,
		Currency: c.Currency{FromCurrencyCode: "USD", ToCurrencyCode: "USD"},
		QuotePrice: c.QuotePrice{
			Price:          110.0,
			PricePrevClose: 100.0,
			PriceOpen:      100.0,
			PriceDayHigh:   110.0,
			PriceDayLow:    90.0,
			Change:         10.0,
			ChangePercent:  10.0,
		},
		QuoteExtended: c.QuoteExtended{
			FiftyTwoWeekHigh: 150,
			FiftyTwoWeekLow:  50,
			MarketCap:        1000000,
		},
		Meta: c.Meta{
			OrderIndex: 0,
		},
	},
	{
		Name:     "Microsoft Inc",
		Symbol:   "MSFT",
		Class:    c.AssetClassStock,
		Currency: c.Currency{FromCurrencyCode: "USD", ToCurrencyCode: "USD"},
		QuotePrice: c.QuotePrice{
			Price:          220.0,
			PricePrevClose: 200.0,
			PriceOpen:      200.0,
			PriceDayHigh:   220.0,
			PriceDayLow:    180.0,
			Change:         20.0,
			ChangePercent:  10.0,
		},
		Meta: c.Meta{
			OrderIndex: 1,
		},
	},
	{
		Name:     "Solana USD",
		Symbol:   "SOL1-USD",
		Class:    c.AssetClassCryptocurrency,
		Currency: c.Currency{FromCurrencyCode: "USD", ToCurrencyCode: "USD"},
		QuotePrice: c.QuotePrice{
			Price:          11.0,
			PricePrevClose: 10.0,
			PriceOpen:      10.0,
			PriceDayHigh:   11.0,
			PriceDayLow:    9.0,
			Change:         1.0,
			ChangePercent:  10.0,
		},
		Meta: c.Meta{
			OrderIndex:          2,
			IsVariablePrecision: true,
		},
	},
}

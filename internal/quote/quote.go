package quote

import (
	c "github.com/achannarasappa/ticker/internal/common"
	quoteYahoo "github.com/achannarasappa/ticker/internal/quote/yahoo"
)

func getQuoteBySource(dep c.Dependencies, symbolBySource c.AssetGroupSymbolsBySource) []c.AssetQuote {

	if symbolBySource.Source == c.QuoteSourceYahoo {
		return quoteYahoo.GetAssetQuotes(*dep.HttpClient, symbolBySource.Symbols)()
	}

	return []c.AssetQuote{}
}

// GetAssetGroupQuote gets price quotes for groups of assets by data source
func GetAssetGroupQuote(dep c.Dependencies) func(c.AssetGroup) c.AssetGroupQuote {

	return func(assetGroup c.AssetGroup) c.AssetGroupQuote {

		var assetQuotes []c.AssetQuote

		for _, symbolBySource := range assetGroup.SymbolsBySource {

			assetQuotebySource := getQuoteBySource(dep, symbolBySource)
			assetQuotes = append(assetQuotes, assetQuotebySource...)

		}

		return c.AssetGroupQuote{
			AssetQuotes: assetQuotes,
			AssetGroup:  assetGroup,
		}
	}
}

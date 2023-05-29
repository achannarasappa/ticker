package quote

import (
	c "github.com/achannarasappa/ticker/internal/common"
	quoteCoingecko "github.com/achannarasappa/ticker/internal/quote/coingecko"
	quoteYahoo "github.com/achannarasappa/ticker/internal/quote/yahoo"
	"github.com/go-resty/resty/v2"
)

func getQuoteBySource(dep c.Dependencies, symbolBySource c.AssetGroupSymbolsBySource) []c.AssetQuote {

	if symbolBySource.Source == c.QuoteSourceYahoo {
		return quoteYahoo.GetAssetQuotes(*dep.HttpClients.Yahoo, symbolBySource.Symbols)()
	}

	if symbolBySource.Source == c.QuoteSourceCoingecko {
		return quoteCoingecko.GetAssetQuotes(*dep.HttpClients.Yahoo, symbolBySource.Symbols)
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

func getUniqueSymbolsBySource(assetGroups []c.AssetGroup) []c.AssetGroupSymbolsBySource {

	symbols := make(map[c.QuoteSource]map[string]bool)
	symbolsUnique := make(map[c.QuoteSource][]string)
	assetGroupSymbolsBySource := make([]c.AssetGroupSymbolsBySource, 0)
	for _, assetGroup := range assetGroups {

		for _, symbolGroup := range assetGroup.SymbolsBySource {

			for _, symbol := range symbolGroup.Symbols {

				source := symbolGroup.Source

				if symbols[source] == nil {
					symbols[source] = map[string]bool{}
				}

				if !symbols[source][symbol] {
					symbols[source][symbol] = true
					symbolsUnique[source] = append(symbolsUnique[source], symbol)
				}
			}

		}

	}

	for source, symbols := range symbolsUnique {
		assetGroupSymbolsBySource = append(assetGroupSymbolsBySource, c.AssetGroupSymbolsBySource{
			Source:  source,
			Symbols: symbols,
		})
	}

	return assetGroupSymbolsBySource

}

// GetAssetGroupsCurrencyRates gets the currency rates by source across all asset groups
func GetAssetGroupsCurrencyRates(client resty.Client, assetGroups []c.AssetGroup, targetCurrency string) (c.CurrencyRates, error) {

	var err error
	var currencyRates c.CurrencyRates
	uniqueSymbolsBySource := getUniqueSymbolsBySource(assetGroups)

	for _, source := range uniqueSymbolsBySource {

		if source.Source == c.QuoteSourceYahoo && err == nil {
			currencyRates, err = quoteYahoo.GetCurrencyRates(client, source.Symbols, targetCurrency)
		}

	}

	return currencyRates, err
}

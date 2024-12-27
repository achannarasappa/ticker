package quote

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	quoteCoinbase "github.com/achannarasappa/ticker/v4/internal/quote/coinbase"
	quoteCoincap "github.com/achannarasappa/ticker/v4/internal/quote/coincap"
	quoteCoingecko "github.com/achannarasappa/ticker/v4/internal/quote/coingecko"
	quoteYahoo "github.com/achannarasappa/ticker/v4/internal/quote/yahoo"
	"github.com/go-resty/resty/v2"
)

func getQuoteBySource(dep c.Dependencies, ref c.Reference, symbolBySource c.AssetGroupSymbolsBySource) []c.AssetQuote {

	if symbolBySource.Source == c.QuoteSourceYahoo {
		return quoteYahoo.GetAssetQuotes(*dep.HttpClients.Yahoo, symbolBySource.Symbols)()
	}

	if symbolBySource.Source == c.QuoteSourceCoingecko {
		return quoteCoingecko.GetAssetQuotes(*dep.HttpClients.Default, symbolBySource.Symbols)
	}

	if symbolBySource.Source == c.QuoteSourceCoinCap {
		return quoteCoincap.GetAssetQuotes(*dep.HttpClients.Default, symbolBySource.Symbols)
	}

	if symbolBySource.Source == c.QuoteSourceCoinbase {

		return quoteCoinbase.GetAssetQuotes(*dep.HttpClients.Default, symbolBySource.Symbols, ref.SourceToUnderlyingAssetSymbols[c.QuoteSourceCoinbase])
	}

	return []c.AssetQuote{}
}

// GetAssetGroupQuote gets price quotes for groups of assets by data source
func GetAssetGroupQuote(dep c.Dependencies, ref c.Reference) func(c.AssetGroup) c.AssetGroupQuote {

	return func(assetGroup c.AssetGroup) c.AssetGroupQuote {

		var assetQuotes []c.AssetQuote

		for _, symbolBySource := range assetGroup.SymbolsBySource {

			assetQuoteBySource := getQuoteBySource(dep, ref, symbolBySource)
			assetQuotes = append(assetQuotes, assetQuoteBySource...)

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
func GetAssetGroupsCurrencyRates(client *resty.Client, assetGroups []c.AssetGroup, targetCurrency string) (c.CurrencyRates, error) {

	var err error
	var currencyRates c.CurrencyRates
	uniqueSymbolsBySource := getUniqueSymbolsBySource(assetGroups)

	for _, source := range uniqueSymbolsBySource {

		if source.Source == c.QuoteSourceYahoo && err == nil {
			currencyRates, err = quoteYahoo.GetCurrencyRates(*client, source.Symbols, targetCurrency)
		}

	}

	return currencyRates, err
}

// GetAssetGroupUnderlyingAssetSymbols retrieves the underlying asset symbol for Coinbase futures contracts that are not already on the watchlist
func GetAssetGroupUnderlyingAssetSymbols(client *resty.Client, assetGroups []c.AssetGroup) (map[c.QuoteSource][]string, error) {
	var err error
	sourceToUnderlyingAssetSymbols := make(map[c.QuoteSource][]string)
	uniqueSymbolsBySource := getUniqueSymbolsBySource(assetGroups)

	for _, source := range uniqueSymbolsBySource {
		if source.Source == c.QuoteSourceCoinbase && err == nil {
			sourceToUnderlyingAssetSymbols[c.QuoteSourceCoinbase], err = quoteCoinbase.GetUnderlyingAssetSymbols(*client, source.Symbols)
		}
	}

	return sourceToUnderlyingAssetSymbols, err

}

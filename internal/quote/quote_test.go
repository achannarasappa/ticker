package quote_test

import (
	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/quote"
	. "github.com/achannarasappa/ticker/test/http"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Quote", func() {

	var (
		dep c.Dependencies
	)

	BeforeEach(func() {
		dep = c.Dependencies{
			HttpClient: client,
		}
		MockResponseYahooQuotes()
	})

	Describe("GetAssetGroupQuote", func() {

		It("should get price quotes for each asset based on it's data source", func() {

			input := c.AssetGroup{
				SymbolsBySource: []c.AssetGroupSymbolsBySource{
					{
						Source: c.QuoteSourceYahoo,
						Symbols: []string{
							"GOOG",
							"RBLX",
						},
					},
					{
						Source: c.QuoteSourceUserDefined,
						Symbols: []string{
							"CASH",
							"PRIVATESHARES",
						},
					},
				},
			}
			output := GetAssetGroupQuote(dep)(input)

			Expect(output.AssetQuotes).To(HaveLen(2))

		})

	})

})

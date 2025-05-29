package api_test

import (
	"os"
	"path/filepath"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	unary "github.com/achannarasappa/ticker/v5/internal/monitor/yahoo/unary"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// Define GetAssetQuotesResponseSchema by loading it from schema.json
var GetAssetQuotesResponseSchema string

func init() {
	// Load the schema file
	schemaBytes, err := os.ReadFile(filepath.Join("schema.json"))
	if err != nil {
		panic("Failed to load schema file: " + err.Error())
	}
	GetAssetQuotesResponseSchema = string(schemaBytes)
}

var _ = Describe("Yahoo API", func() {
	Describe("GetAssetQuotes Response", func() {
		It("should have expected fields in the response", func() {
			// Setup API client
			client := unary.NewUnaryAPI(unary.Config{
				BaseURL:           "https://query1.finance.yahoo.com",
				SessionRootURL:    "https://finance.yahoo.com",
				SessionCrumbURL:   "https://query2.finance.yahoo.com",
				SessionConsentURL: "https://consent.yahoo.com",
			})

			// Get quotes using the client API
			quotes, _, err := client.GetAssetQuotes([]string{"AAPL"})
			Expect(err).NotTo(HaveOccurred())
			Expect(quotes).NotTo(BeEmpty())

			// Get the first quote for AAPL
			quote := quotes[0]

			// Validate the quote structure matches the AssetQuote properties
			// Basic properties
			Expect(quote.Symbol).To(Equal("AAPL"))
			Expect(quote.Name).To(ContainSubstring("Apple"))

			// Asset class
			Expect(quote.Class).To(BeNumerically(">=", 0))

			// Currency
			Expect(quote.Currency.FromCurrencyCode).NotTo(BeEmpty())

			// QuotePrice properties
			Expect(quote.QuotePrice.Price).NotTo(BeZero())
			Expect(quote.QuotePrice.PricePrevClose).NotTo(BeZero())
			Expect(quote.QuotePrice.PriceOpen).NotTo(BeZero())
			Expect(quote.QuotePrice.PriceDayHigh).NotTo(BeZero())
			Expect(quote.QuotePrice.PriceDayLow).NotTo(BeZero())
			Expect(quote.QuotePrice.Change).To(BeAssignableToTypeOf(0.0))
			Expect(quote.QuotePrice.ChangePercent).To(BeAssignableToTypeOf(0.0))

			// QuoteExtended properties
			Expect(quote.QuoteExtended.FiftyTwoWeekHigh).NotTo(BeZero())
			Expect(quote.QuoteExtended.FiftyTwoWeekLow).NotTo(BeZero())
			Expect(quote.QuoteExtended.MarketCap).NotTo(BeZero())
			Expect(quote.QuoteExtended.Volume).NotTo(BeZero())

			// Exchange properties
			Expect(quote.Exchange.Name).NotTo(BeEmpty())
			Expect(quote.Exchange.Delay).To(BeNumerically(">=", 0))
			Expect(quote.Exchange.State).To(BeNumerically(">=", 0))

			// QuoteSource
			Expect(quote.QuoteSource).To(Equal(c.QuoteSourceYahoo))

			// Meta
			Expect(quote.Meta.SymbolInSourceAPI).To(Equal("AAPL"))

			// Check that PostMarket and PreMarket are conditionally present
			// We only validate them if they're present, as they depend on market times
			if quote.Exchange.State == 2 { // PostMarket
				Expect(quote.QuotePrice.Change).NotTo(BeZero())
				Expect(quote.QuotePrice.ChangePercent).NotTo(BeZero())
				Expect(quote.QuotePrice.Price).NotTo(BeZero())
			}

			if quote.Exchange.State == 1 { // PreMarket
				Expect(quote.QuotePrice.Change).NotTo(BeZero())
				Expect(quote.QuotePrice.ChangePercent).NotTo(BeZero())
				Expect(quote.QuotePrice.Price).NotTo(BeZero())
			}
		})
	})
})

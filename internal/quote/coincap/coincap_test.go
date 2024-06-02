package coincap_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/quote/coincap"
	. "github.com/achannarasappa/ticker/v4/test/http"
	g "github.com/onsi/gomega/gstruct"
)

var _ = Describe("CoinCap Quote", func() {
	Describe("GetAssetQuotes", func() {
		It("should make a request to get stock quotes and transform the response", func() {

			MockResponseCoincapQuotes()

			output := GetAssetQuotes(*client, []string{"elrond"})
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(63420.89053262873),
						"PricePrevClose": Equal(64284.8148182606),
						"PriceOpen":      Equal(0.0),
						"PriceDayHigh":   Equal(0.0),
						"PriceDayLow":    Equal(0.0),
						"Change":         BeNumerically("~", 863.9242856318742, 1e8),
						"ChangePercent":  Equal(1.3622077494913285),
					}),
					"QuoteSource": Equal(c.QuoteSourceCoinCap),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
				}),
			}))
		})
	})
})

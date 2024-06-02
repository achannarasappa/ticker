package coingecko_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/quote/coingecko"
	. "github.com/achannarasappa/ticker/v4/test/http"
	g "github.com/onsi/gomega/gstruct"
)

var _ = Describe("Coingecko", func() {

	Describe("GetAssetQuotes", func() {

		It("should make a request to get crypto quotes and transform the response", func() {
			MockResponseCoingeckoQuotes()

			output := GetAssetQuotes(*client, []string{"bitcoin"})
			Expect(output).To(g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
				"0": g.MatchFields(g.IgnoreExtras, g.Fields{
					"QuotePrice": g.MatchFields(g.IgnoreExtras, g.Fields{
						"Price":          Equal(39045.0),
						"PricePrevClose": Equal(40023.048909591314),
						"PriceOpen":      Equal(0.0),
						"PriceDayHigh":   Equal(40090.0),
						"PriceDayLow":    Equal(38195.0),
						"Change":         Equal(-978.048909591315),
						"ChangePercent":  Equal(-2.44373),
					}),
					"QuoteSource": Equal(c.QuoteSourceCoingecko),
					"Exchange": g.MatchFields(g.IgnoreExtras, g.Fields{
						"IsActive":                BeTrue(),
						"IsRegularTradingSession": BeTrue(),
					}),
				}),
			}))
		})

	})

})

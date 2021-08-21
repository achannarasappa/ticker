package asset_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/internal/asset"
	c "github.com/achannarasappa/ticker/internal/common"
)

var _ = Describe("Asset", func() {

	Describe("GetAssets", func() {
		It("should return assets", func() {
			inputContext := c.Context{}
			inputAssetQuotes := fixtureAssetQuotes

			expectedAssets := fixtureAssets
			expectedHoldingSummary := HoldingSummary{}

			outputAssets, outputHoldingSummary := GetAssets(inputContext, inputAssetQuotes)

			Expect(outputAssets).To(Equal(expectedAssets))
			Expect(outputHoldingSummary).To(Equal(expectedHoldingSummary))
		})

		When("there are lots", func() {
			It("should return assets with holdings and a summary of holdings", func() {
				inputContext := c.Context{}
				inputAssetQuotes := fixtureAssetQuotes
				inputContext.Config.Lots = []c.Lot{
					{
						Symbol:    "TWKS",
						UnitCost:  100,
						Quantity:  10,
						FixedCost: 7,
					},
					{
						Symbol:    "TWKS",
						UnitCost:  75,
						Quantity:  10,
						FixedCost: 0,
					},
					{
						Symbol:    "MSFT",
						UnitCost:  400,
						Quantity:  10,
						FixedCost: 0,
					},
				}

				outputAssets, outputHoldingSummary := GetAssets(inputContext, inputAssetQuotes)

				Expect(outputAssets).To(HaveLen(3))

				Expect(outputAssets[0].Holding.Value).To(Equal(2200.0))
				Expect(outputAssets[0].Holding.Cost).To(Equal(1757.0))
				Expect(outputAssets[0].Holding.Quantity).To(Equal(20.0))
				Expect(outputAssets[0].Holding.UnitValue).To(Equal(110.0))
				Expect(outputAssets[0].Holding.UnitCost).To(Equal(87.85))
				Expect(outputAssets[0].Holding.Weight).To(Equal(50.0))

				Expect(outputAssets[1].Holding.Value).To(Equal(2200.0))
				Expect(outputAssets[1].Holding.Cost).To(Equal(4000.0))
				Expect(outputAssets[1].Holding.Quantity).To(Equal(10.0))
				Expect(outputAssets[1].Holding.UnitValue).To(Equal(220.0))
				Expect(outputAssets[1].Holding.UnitCost).To(Equal(400.0))
				Expect(outputAssets[1].Holding.Weight).To(Equal(50.0))

				Expect(outputAssets[2].Holding).To(Equal(c.Holding{}))

				Expect(outputHoldingSummary.Cost).To(Equal(5757.0))
				Expect(outputHoldingSummary.Value).To(Equal(4400.0))
				Expect(outputHoldingSummary.TotalChange.Amount).To(Equal(-1357.0))
				Expect(outputHoldingSummary.TotalChange.Percent).To(Equal(-23.57130449887094))
				Expect(outputHoldingSummary.DayChange.Amount).To(Equal(400.0))
				Expect(outputHoldingSummary.DayChange.Percent).To(Equal(9.090909090909092))
			})
		})

		When("there is a currency to convert", func() {

			inputContext := c.Context{
				Config: c.Config{
					Currency: "EUR",
				},
				Reference: c.Reference{
					CurrencyRates: map[string]c.CurrencyRate{
						"USD": {
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         1.5,
						},
					},
				},
			}
			inputAssetQuotes := fixtureAssetQuotes
			inputContext.Config.Lots = []c.Lot{
				{
					Symbol:    "TWKS",
					UnitCost:  100,
					Quantity:  10,
					FixedCost: 0,
				},
			}

			outputAssets, outputHoldingSummary := GetAssets(inputContext, inputAssetQuotes)

			It("should set currency from and to when converting", func() {
				Expect(outputAssets[0].Currency.FromCurrencyCode).To(Equal("USD"))
				Expect(outputAssets[0].Currency.ToCurrencyCode).To(Equal("EUR"))
			})
			It("should convert the currencies for quote", func() {
				Expect(outputAssets[0].QuotePrice.Price).To(Equal(165.0))
				Expect(outputAssets[0].QuotePrice.PricePrevClose).To(Equal(150.0))
				Expect(outputAssets[0].QuotePrice.PriceOpen).To(Equal(150.0))
				Expect(outputAssets[0].QuotePrice.PriceDayHigh).To(Equal(165.0))
				Expect(outputAssets[0].QuotePrice.PriceDayLow).To(Equal(135.0))
				Expect(outputAssets[0].QuotePrice.Change).To(Equal(15.0))
			})
			It("should convert the currencies for quote extended", func() {
				Expect(outputAssets[0].QuoteExtended.FiftyTwoWeekHigh).To(Equal(225.0))
				Expect(outputAssets[0].QuoteExtended.FiftyTwoWeekLow).To(Equal(75.0))
				Expect(outputAssets[0].QuoteExtended.MarketCap).To(Equal(1500000.0))
			})
			It("should convert holding values", func() {
				Expect(outputAssets[0].Holding.Value).To(Equal(1650.0))
				Expect(outputAssets[0].Holding.Cost).To(Equal(1500.0))
				Expect(outputAssets[0].Holding.UnitValue).To(Equal(165.0))
				Expect(outputAssets[0].Holding.UnitCost).To(Equal(150.0))
			})
			It("should convert holding summary values", func() {
				Expect(outputHoldingSummary.Cost).To(Equal(1500.0))
				Expect(outputHoldingSummary.Value).To(Equal(1650.0))
				Expect(outputHoldingSummary.TotalChange.Amount).To(Equal(150.0))
				Expect(outputHoldingSummary.DayChange.Amount).To(Equal(150.0))
			})
		})

		When("there is no explicit currency conversion", func() {
			When("and there is a holding with a non-US currency", func() {

				inputContext := c.Context{
					Reference: c.Reference{
						CurrencyRates: map[string]c.CurrencyRate{
							"EUR": {
								FromCurrency: "EUR",
								ToCurrency:   "USD",
								Rate:         0.5,
							},
						},
					},
				}
				inputAssetQuotes := make([]c.AssetQuote, len(fixtureAssetQuotes))
				copy(inputAssetQuotes, fixtureAssetQuotes)
				inputAssetQuotes[0].Currency.FromCurrencyCode = "EUR"
				inputContext.Config.Lots = []c.Lot{
					{
						Symbol:    "TWKS",
						UnitCost:  100,
						Quantity:  10,
						FixedCost: 0,
					},
				}

				outputAssets, outputHoldingSummary := GetAssets(inputContext, inputAssetQuotes)

				It("should convert the currency for the holding summary", func() {
					Expect(outputHoldingSummary.Value).To(Equal(550.0))
					Expect(outputHoldingSummary.Cost).To(Equal(500.0))
				})
				It("should not convert the currencies for quote", func() {
					Expect(outputAssets[0].QuotePrice.Price).To(Equal(110.0))
					Expect(outputAssets[0].QuoteExtended.MarketCap).To(Equal(1000000.0))
				})
				It("should not convert the currency for the holding", func() {
					Expect(outputAssets[0].Holding.Value).To(Equal(1100.0))
					Expect(outputAssets[0].Holding.Cost).To(Equal(1000.0))
				})
			})
		})

		When("there are no asset quotes", func() {
			It("should return empty assets and holdings summary", func() {
				outputAssets, outputHoldingSummary := GetAssets(c.Context{}, []c.AssetQuote{})

				Expect(outputAssets).To(Equal([]c.Asset{}))
				Expect(outputHoldingSummary).To(Equal(HoldingSummary{}))
			})
		})
	})

	Describe("GetSymbols", func() {

		It("should return a slice of symbols", func() {

			inputConfig := c.Config{
				Watchlist: []string{"GOOG", "ARKW"},
				Lots: []c.Lot{
					{Symbol: "ABNB", UnitCost: 100, Quantity: 10},
					{Symbol: "ARKW", UnitCost: 200, Quantity: 10},
				},
				ShowHoldings: true,
			}
			output := GetSymbols(inputConfig)
			expected := []string{
				"ABNB",
				"ARKW",
				"GOOG",
			}
			Expect(output).To(ContainElements(expected))
		})

		When("holdings are hidden", func() {
			It("should not show symbols for holdings", func() {
				inputConfig := c.Config{
					Watchlist:    []string{"GOOG", "ARKW"},
					ShowHoldings: false,
				}
				output := GetSymbols(inputConfig)
				expected := []string{
					"ARKW",
					"GOOG",
				}
				Expect(output).To(ContainElements(expected))
			})
		})
	})
})

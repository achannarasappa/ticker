package asset_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/v5/internal/asset"
	c "github.com/achannarasappa/ticker/v5/internal/common"
)

var _ = Describe("Asset", func() {

	Describe("GetAssets", func() {
		It("should return assets", func() {
			inputContext := c.Context{}
			inputAssetGroupQuote := fixtureAssetGroupQuote

			expectedAssets := fixtureAssets
			expectedPositionSummary := PositionSummary{}

			outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

			Expect(outputAssets).To(Equal(expectedAssets))
			Expect(outputPositionSummary).To(Equal(expectedPositionSummary))
		})

		When("there are lots", func() {
			It("should return assets with positions and a summary of positions", func() {
				inputContext := c.Context{}
				inputAssetGroupQuote := fixtureAssetGroupQuote
				inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
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

				outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

				Expect(outputAssets).To(HaveLen(3))

				Expect(outputAssets[0].Position.Value).To(Equal(2200.0))
				Expect(outputAssets[0].Position.Cost).To(Equal(1757.0))
				Expect(outputAssets[0].Position.Quantity).To(Equal(20.0))
				Expect(outputAssets[0].Position.UnitValue).To(Equal(110.0))
				Expect(outputAssets[0].Position.UnitCost).To(Equal(87.85))
				Expect(outputAssets[0].Position.Weight).To(Equal(50.0))

				Expect(outputAssets[1].Position.Value).To(Equal(2200.0))
				Expect(outputAssets[1].Position.Cost).To(Equal(4000.0))
				Expect(outputAssets[1].Position.Quantity).To(Equal(10.0))
				Expect(outputAssets[1].Position.UnitValue).To(Equal(220.0))
				Expect(outputAssets[1].Position.UnitCost).To(Equal(400.0))
				Expect(outputAssets[1].Position.Weight).To(Equal(50.0))

				Expect(outputAssets[2].Position).To(Equal(c.Position{}))

				Expect(outputPositionSummary.Cost).To(Equal(5757.0))
				Expect(outputPositionSummary.Value).To(Equal(4400.0))
				Expect(outputPositionSummary.TotalChange.Amount).To(Equal(-1357.0))
				Expect(outputPositionSummary.TotalChange.Percent).To(Equal(-23.57130449887094))
				Expect(outputPositionSummary.DayChange.Amount).To(Equal(400.0))
				Expect(outputPositionSummary.DayChange.Percent).To(Equal(9.090909090909092))
			})
		})

		When("there is a currency to convert", func() {

			inputContext := c.Context{
				Config: c.Config{
					Currency: "EUR",
				},
				Reference: c.Reference{},
			}
			inputAssetGroupQuote := fixtureAssetGroupQuote
			inputAssetGroupQuote.AssetQuotes = []c.AssetQuote{
				{
					Name:          "ThoughtWorks",
					Symbol:        "TWKS",
					Class:         c.AssetClassStock,
					Currency:      c.Currency{FromCurrencyCode: "USD", ToCurrencyCode: "EUR", Rate: 1.5},
					QuotePrice:    c.QuotePrice{Price: 110.0, PricePrevClose: 100.0, PriceOpen: 100.0, PriceDayHigh: 110.0, PriceDayLow: 90.0, Change: 10.0, ChangePercent: 10.0},
					QuoteExtended: c.QuoteExtended{FiftyTwoWeekHigh: 150, FiftyTwoWeekLow: 50, MarketCap: 1000000},
				},
				{
					Name:       "Microsoft Inc",
					Symbol:     "MSFT",
					Class:      c.AssetClassStock,
					Currency:   c.Currency{FromCurrencyCode: "GBP", ToCurrencyCode: "EUR", Rate: 2.0},
					QuotePrice: c.QuotePrice{Price: 220.0, PricePrevClose: 200.0, PriceOpen: 200.0, PriceDayHigh: 220.0, PriceDayLow: 180.0, Change: 20.0, ChangePercent: 10.0},
				},
			}
			inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
				{
					Symbol:    "TWKS",
					UnitCost:  100,
					Quantity:  10,
					FixedCost: 0,
				},
				{
					Symbol:    "MSFT",
					UnitCost:  100,
					Quantity:  1,
					FixedCost: 0,
				},
			}

			outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

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
			It("should convert position values", func() {
				Expect(outputAssets[0].Position.Value).To(Equal(1650.0))
				Expect(outputAssets[0].Position.Cost).To(Equal(1500.0))
				Expect(outputAssets[0].Position.UnitValue).To(Equal(165.0))
				Expect(outputAssets[0].Position.UnitCost).To(Equal(150.0))
			})
			It("should convert position summary values", func() {
				Expect(outputPositionSummary.Cost).To(Equal(1700.0))
				Expect(outputPositionSummary.Value).To(Equal(2090.0))
				Expect(outputPositionSummary.TotalChange.Amount).To(Equal(390.0))
				Expect(outputPositionSummary.DayChange.Amount).To(Equal(190.0))
			})

			When("and the summary conversion only option is set", func() {

				inputContextSummaryConversion := inputContext
				inputContextSummaryConversion.Config = c.Config{
					Currency:                   "EUR",
					CurrencyConvertSummaryOnly: true,
				}

				_, outputPositionSummary := GetAssets(inputContextSummaryConversion, inputAssetGroupQuote)

				It("should convert position summary values", func() {
					Expect(outputPositionSummary.Cost).To(Equal(1700.0))
					Expect(outputPositionSummary.Value).To(Equal(2090.0))
					Expect(outputPositionSummary.TotalChange.Amount).To(Equal(390.0))
					Expect(outputPositionSummary.DayChange.Amount).To(Equal(190.0))
				})

			})

			When("and the disable unit cost conversion option is set", func() {

				inputContextDisableUnitCostConversion := inputContext
				inputContextDisableUnitCostConversion.Config = c.Config{
					Currency:                          "EUR",
					CurrencyDisableUnitCostConversion: true,
				}

				outputAssets, outputPositionSummary := GetAssets(inputContextDisableUnitCostConversion, inputAssetGroupQuote)

				It("should not convert position costs", func() {
					Expect(outputAssets[0].Position.Cost).To(Equal(1000.0)) // 1000 EUR unconverted since option is set
					Expect(outputAssets[0].Position.UnitCost).To(Equal(100.0))
					Expect(outputAssets[0].Position.Value).To(Equal(1650.0)) // Conversion 10 shares @ 110 USD/share to EUR
					Expect(outputAssets[0].Position.TotalChange.Amount).To(Equal(650.0))
					Expect(outputAssets[0].Position.TotalChange.Percent).To(Equal(65.0))
					Expect(outputPositionSummary.Cost).To(Equal(1100.0)) // Sum of 1000 EUR + 100 EUR
					Expect(outputPositionSummary.DayChange.Percent).To(Equal(9.090909090909092))
					Expect(outputPositionSummary.TotalChange.Amount).To(Equal(990.0))
					Expect(outputPositionSummary.TotalChange.Percent).To(Equal(90.0))
				})

			})

		})

		When("there is no explicit currency conversion", func() {
			When("and there is a position with a non-US currency", func() {

				inputContext := c.Context{
					Reference: c.Reference{},
				}
				inputAssetGroupQuote := fixtureAssetGroupQuote
				inputAssetQuotes := make([]c.AssetQuote, len(fixtureAssetGroupQuote.AssetQuotes))
				copy(inputAssetQuotes, fixtureAssetGroupQuote.AssetQuotes)
				inputAssetQuotes[0].Currency.FromCurrencyCode = "EUR"
				inputAssetQuotes[0].Currency.ToCurrencyCode = "USD"
				inputAssetQuotes[0].Currency.Rate = 0.5
				inputAssetGroupQuote.AssetQuotes = inputAssetQuotes
				inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
					{
						Symbol:    "TWKS",
						UnitCost:  100,
						Quantity:  10,
						FixedCost: 0,
					},
				}

				outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

				It("should convert the currency for the position summary", func() {
					Expect(outputPositionSummary.Value).To(Equal(550.0))
					Expect(outputPositionSummary.Cost).To(Equal(500.0))
				})
				It("should not convert the currencies for quote", func() {
					Expect(outputAssets[0].QuotePrice.Price).To(Equal(110.0))
					Expect(outputAssets[0].QuoteExtended.MarketCap).To(Equal(1000000.0))
				})
				It("should not convert the currency for the position", func() {
					Expect(outputAssets[0].Position.Value).To(Equal(1100.0))
					Expect(outputAssets[0].Position.Cost).To(Equal(1000.0))
				})
			})
		})

		When("there are no asset quotes", func() {
			It("should return empty assets and positions summary", func() {
				outputAssets, outputPositionSummary := GetAssets(c.Context{}, c.AssetGroupQuote{})

				Expect(outputAssets).To(Equal([]c.Asset{}))
				Expect(outputPositionSummary).To(Equal(PositionSummary{}))
			})
		})

		When("unit cost is zero", func() {
			When("and fixed cost is also zero", func() {
				It("should handle zero cost without division by zero", func() {
					inputContext := c.Context{}
					inputAssetGroupQuote := fixtureAssetGroupQuote
					inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
						{
							Symbol:    "TWKS",
							UnitCost:  0,
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

					outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

					Expect(outputAssets).To(HaveLen(3))

					// TWKS with zero cost
					Expect(outputAssets[0].Position.Value).To(Equal(1100.0)) // 10 * 110
					Expect(outputAssets[0].Position.Cost).To(Equal(0.0))
					Expect(outputAssets[0].Position.Quantity).To(Equal(10.0))
					Expect(outputAssets[0].Position.UnitValue).To(Equal(110.0))
					Expect(outputAssets[0].Position.UnitCost).To(Equal(0.0))
					Expect(outputAssets[0].Position.TotalChange.Amount).To(Equal(1100.0)) // value - cost
					Expect(outputAssets[0].Position.TotalChange.Percent).To(Equal(0.0))   // Should be 0 when cost is 0

					// MSFT with normal cost
					Expect(outputAssets[1].Position.Value).To(Equal(2200.0))
					Expect(outputAssets[1].Position.Cost).To(Equal(4000.0))
					Expect(outputAssets[1].Position.Quantity).To(Equal(10.0))
					Expect(outputAssets[1].Position.UnitValue).To(Equal(220.0))
					Expect(outputAssets[1].Position.UnitCost).To(Equal(400.0))

					// Summary should include zero cost positions
					Expect(outputPositionSummary.Cost).To(Equal(4000.0))
					Expect(outputPositionSummary.Value).To(Equal(3300.0)) // 1100 + 2200
					Expect(outputPositionSummary.TotalChange.Amount).To(Equal(-700.0)) // 3300 - 4000
					Expect(outputPositionSummary.TotalChange.Percent).To(Equal(-17.5)) // (-700 / 4000) * 100
				})
			})

			When("but fixed cost is non-zero", func() {
				It("should use fixed cost as the total cost", func() {
					inputContext := c.Context{}
					inputAssetGroupQuote := fixtureAssetGroupQuote
					inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
						{
							Symbol:    "TWKS",
							UnitCost:  0,
							Quantity:  10,
							FixedCost: 50,
						},
					}

					outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

					Expect(outputAssets).To(HaveLen(3))

					Expect(outputAssets[0].Position.Value).To(Equal(1100.0)) // 10 * 110
					Expect(outputAssets[0].Position.Cost).To(Equal(50.0))    // Only fixed cost
					Expect(outputAssets[0].Position.Quantity).To(Equal(10.0))
					Expect(outputAssets[0].Position.UnitValue).To(Equal(110.0))
					Expect(outputAssets[0].Position.UnitCost).To(Equal(5.0)) // 50 / 10
					Expect(outputAssets[0].Position.TotalChange.Amount).To(Equal(1050.0))
					Expect(outputAssets[0].Position.TotalChange.Percent).To(Equal(2100.0)) // (1050 / 50) * 100

					Expect(outputPositionSummary.Cost).To(Equal(50.0))
					Expect(outputPositionSummary.Value).To(Equal(1100.0))
					Expect(outputPositionSummary.TotalChange.Amount).To(Equal(1050.0))
					Expect(outputPositionSummary.TotalChange.Percent).To(Equal(2100.0))
				})
			})
		})

		When("unit cost is undefined (defaults to zero)", func() {
			It("should handle undefined unit cost the same as zero", func() {
				inputContext := c.Context{}
				inputAssetGroupQuote := fixtureAssetGroupQuote
				inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
					{
						Symbol:    "TWKS",
						Quantity:  10,
						FixedCost: 0,
						// UnitCost is not set, defaults to 0
					},
				}

				outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

				Expect(outputAssets).To(HaveLen(3))

				Expect(outputAssets[0].Position.Value).To(Equal(1100.0))
				Expect(outputAssets[0].Position.Cost).To(Equal(0.0))
				Expect(outputAssets[0].Position.UnitCost).To(Equal(0.0))
				Expect(outputAssets[0].Position.TotalChange.Percent).To(Equal(0.0)) // Should not cause division by zero

				// Summary should include zero cost positions
				Expect(outputPositionSummary.Cost).To(Equal(0.0))
				Expect(outputPositionSummary.Value).To(Equal(1100.0))
				Expect(outputPositionSummary.TotalChange.Amount).To(Equal(1100.0)) // 1100 - 0
				Expect(outputPositionSummary.TotalChange.Percent).To(Equal(0.0)) // 0 when cost is 0
			})
		})

		When("aggregated quantity is zero", func() {
			It("should handle a net zero position", func() {
				inputContext := c.Context{}
				inputAssetGroupQuote := fixtureAssetGroupQuote
				inputAssetGroupQuote.AssetGroup.ConfigAssetGroup.Lots = []c.Lot{
					{
						Symbol:    "TWKS",
						UnitCost:  100.0,
						Quantity:  10.0,
						FixedCost: 0,
					},
					{
						Symbol:    "TWKS",
						UnitCost:  110.0,
						Quantity:  -10.0,
						FixedCost: 0,
					},
				}

				outputAssets, outputPositionSummary := GetAssets(inputContext, inputAssetGroupQuote)

				Expect(outputAssets).To(HaveLen(3))

				// TWKS with zero aggregated quantity (10 + (-10) = 0)
				// Cost calculation: (100*10) + (110*(-10)) = 1000 - 1100 = -100
				Expect(outputAssets[0].Position.Value).To(Equal(0.0))    // 0 * 110 = 0
				Expect(outputAssets[0].Position.Cost).To(Equal(-100.0)) // Aggregated cost: 1000 + (-1100) = -100
				Expect(outputAssets[0].Position.Quantity).To(Equal(0.0)) // 10 + (-10) = 0
				Expect(outputAssets[0].Position.UnitValue).To(Equal(0.0)) // Should be 0
				Expect(outputAssets[0].Position.UnitCost).To(Equal(0.0))  // Should be 0
				Expect(outputAssets[0].Position.TotalChange.Amount).To(Equal(100.0))   // value - cost = 0 - (-100) = 100
				Expect(outputAssets[0].Position.TotalChange.Percent).To(Equal(-100.0)) // (100 / -100) * 100 = -100

				// Summary should handle zero quantity positions
				// Since position.Value is 0, it's skipped in addHoldingToPositionSummary
				Expect(outputPositionSummary.Cost).To(Equal(0.0))
				Expect(outputPositionSummary.Value).To(Equal(0.0))
				Expect(outputPositionSummary.TotalChange.Amount).To(Equal(0.0))
			})
		})

	})
})

package asset_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "github.com/achannarasappa/ticker/v5/internal/asset"
	c "github.com/achannarasappa/ticker/v5/internal/common"
)

var assetGroupFixture = c.AssetGroup{
	ConfigAssetGroup: c.ConfigAssetGroup{
		Lots: []c.Lot{
			{
				Symbol:   "TEST1",
				Quantity: 10,
				UnitCost: 100.0,
			},
			{
				Symbol:   "TEST2",
				Quantity: 10,
				UnitCost: 50.0,
			},
		},
	},
}

var _ = Describe("Currency", func() {

	Describe("Currency conversion through GetAssets", func() {
		When("currency conversion is not enabled", func() {
			When("there is no currency conversion rate available", func() {
				It("should not convert currency", func() {
					inputCtx := c.Context{
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: assetGroupFixture,
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "TEST1",
								Currency: c.Currency{
									FromCurrencyCode: "EUR",
									ToCurrencyCode:   "USD",
									Rate:             0,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
							{
								Symbol: "TEST2",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "USD",
									Rate:             0,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					Expect(assets[0].Currency.FromCurrencyCode).To(Equal("EUR"))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("EUR"))
					Expect(assets[0].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[0].Position.Value).To(Equal(1000.0))  // No conversion applied
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[1].Position.Value).To(Equal(1000.0))  // No conversion applied
					// These may be incorrect (only if differing quoting currencies) since different currencies may be added together however if currency rate is not available, nothing can be done to fix
					Expect(summary.Value).To(Equal(2000.0))
					Expect(summary.Cost).To(Equal(1500.0))
				})
			})

			When("there is a currency conversion rate available", func() {
				It("should not convert currency", func() {
					inputCtx := c.Context{
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: assetGroupFixture,
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "TEST1",
								Currency: c.Currency{
									FromCurrencyCode: "EUR",
									ToCurrencyCode:   "USD",
									Rate:             4,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
							{
								Symbol: "TEST2",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "USD",
									Rate:             8,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					Expect(assets[0].Currency.FromCurrencyCode).To(Equal("EUR"))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("USD"))
					Expect(assets[0].QuotePrice.Price).To(Equal(100.0))
					Expect(assets[0].Position.Value).To(Equal(1000.0))
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("USD"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0))
					Expect(assets[1].Position.Value).To(Equal(1000.0))
					// These will be correct since there is a conversion rate available to USD
					Expect(summary.Value).To(Equal(12000.0))
					Expect(summary.Cost).To(Equal(8000.0))
				})
			})
		})

		When("currency conversion is enabled", func() {
			When("there is a currency conversion rate set", func() {
				It("should apply the conversion rate to all values", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency: "EUR",
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: assetGroupFixture,
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "TEST1",
								Currency: c.Currency{
									FromCurrencyCode: "USD",
									ToCurrencyCode:   "EUR",
									Rate:             1.25,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
							{
								Symbol: "TEST2",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "EUR",
									Rate:             0.50,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					Expect(assets[0].Currency.FromCurrencyCode).To(Equal("USD"))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("EUR"))
					Expect(assets[0].QuotePrice.Price).To(Equal(125.0))
					Expect(assets[0].Position.Cost).To(Equal(1250.0))
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("EUR"))
					Expect(assets[1].QuotePrice.Price).To(Equal(50.0))
					Expect(assets[1].Position.Cost).To(Equal(250.0))
					Expect(summary.Value).To(Equal(1750.0))
					Expect(summary.Cost).To(Equal(1500.0))
				})
			})

			When("there is no currency conversion rate set", func() {
				It("should not convert currency since none is available", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency: "EUR",
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: assetGroupFixture,
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "TEST1",
								Currency: c.Currency{
									FromCurrencyCode: "USD",
									ToCurrencyCode:   "EUR",
									Rate:             0,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
							{
								Symbol: "TEST2",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "EUR",
									Rate:             0,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					Expect(assets[0].Currency.FromCurrencyCode).To(Equal("USD"))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("USD"))
					Expect(assets[0].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[0].Position.Value).To(Equal(1000.0))  // No conversion applied
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[1].Position.Value).To(Equal(1000.0))  // No conversion applied
					// These may be incorrect (only if differing quoting currencies) since different currencies may be added together however if currency rate is not available, nothing can be done to fix
					Expect(summary.Value).To(Equal(2000.0))
					Expect(summary.Cost).To(Equal(1500.0))
				})
			})

			When("unit cost conversion is disabled", func() {
				It("should not convert unit costs or summary costs", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency: "EUR",
							// This setting can be used to avoid converting costs in cases where the unit cost is already in the above specified currency
							CurrencyDisableUnitCostConversion: true,
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: assetGroupFixture,
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "TEST1",
								Currency: c.Currency{
									FromCurrencyCode: "USD",
									ToCurrencyCode:   "EUR",
									Rate:             1.25,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
							{
								Symbol: "TEST2",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "EUR",
									Rate:             0.50,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					Expect(assets[0].Currency.FromCurrencyCode).To(Equal("USD"))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("EUR"))
					Expect(assets[0].QuotePrice.Price).To(Equal(125.0)) // Converted
					Expect(assets[0].Position.Cost).To(Equal(1000.0))   // Not converted
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("EUR"))
					Expect(assets[1].QuotePrice.Price).To(Equal(50.0)) // Converted
					Expect(assets[1].Position.Cost).To(Equal(500.0))   // Not converted
					Expect(summary.Value).To(Equal(1750.0))            // Converted
					Expect(summary.Cost).To(Equal(1500.0))             // Not converted
				})
			})

			When("summary only conversion is enabled", func() {
				It("should convert quote prices but not summary values", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency:                   "EUR",
							CurrencyConvertSummaryOnly: true,
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: assetGroupFixture,
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "TEST1",
								Currency: c.Currency{
									FromCurrencyCode: "USD",
									ToCurrencyCode:   "EUR",
									Rate:             1.25,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
							{
								Symbol: "TEST2",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "EUR",
									Rate:             0.50,
								},
								QuotePrice: c.QuotePrice{
									Price: 100.0,
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					Expect(assets[0].Currency.FromCurrencyCode).To(Equal("USD"))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("USD"))
					Expect(assets[0].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[0].Position.Value).To(Equal(1000.0))  // No conversion applied
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[1].Position.Value).To(Equal(1000.0))  // No conversion applied
					Expect(summary.Value).To(Equal(1750.0))             // Converted
					Expect(summary.Cost).To(Equal(1500.0))              // Converted
				})
			})

		})

		When("a currency pair is explicitly watched", func() {
			It("should not convert currency even when currency conversion is enabled", func() {
				inputCtx := c.Context{
					Config: c.Config{
						Currency: "EUR",
					},
					Reference: c.Reference{},
				}
				assetGroupQuote := c.AssetGroupQuote{
					AssetGroup: c.AssetGroup{
						ConfigAssetGroup: c.ConfigAssetGroup{
							Lots: []c.Lot{
								{
									Symbol:   "KRW=X",
									Quantity: 1000,
									UnitCost: 0.00075, // Cost in USD
								},
							},
						},
					},
					AssetQuotes: []c.AssetQuote{
						{
							Symbol: "KRW=X",
							Class:  c.AssetClassCurrency,
							Currency: c.Currency{
								FromCurrencyCode: "USD",
								ToCurrencyCode:   "EUR",
								Rate:             1.25, // This rate should be ignored for currency pairs
							},
							QuotePrice: c.QuotePrice{
								Price: 0.00075, // Exchange rate: 1 KRW = 0.00075 USD
							},
						},
					},
				}
				assets, summary := GetAssets(inputCtx, assetGroupQuote)

				Expect(assets).To(HaveLen(1))
				Expect(assets[0].Currency.FromCurrencyCode).To(Equal("USD"))
				Expect(assets[0].Currency.ToCurrencyCode).To(Equal("USD")) // Should remain USD, not converted to EUR
				Expect(assets[0].QuotePrice.Price).To(Equal(0.00075))      // No conversion applied
				Expect(assets[0].Position.Value).To(Equal(0.75))           // 1000 * 0.00075, no conversion
				Expect(assets[0].Position.Cost).To(Equal(0.75))            // No conversion applied
				Expect(summary.Value).To(Equal(0.75))                      // No conversion applied
				Expect(summary.Cost).To(Equal(0.75))                       // No conversion applied
			})
		})

		Describe("Position weight with minor currency assets", func() {
			When("a single asset is quoted in a minor currency and no display currency is set", func() {
				It("should weight the position at 100% by converting the value to the summary currency before dividing", func() {
					inputCtx := c.Context{
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: c.AssetGroup{
							ConfigAssetGroup: c.ConfigAssetGroup{
								Lots: []c.Lot{
									{
										Symbol:   "JEGP.L",
										Quantity: 10,
										UnitCost: 4000.0, // GBp (quote currency)
									},
								},
							},
						},
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "JEGP.L",
								Currency: c.Currency{
									FromCurrencyCode: "GBp",
									ToCurrencyCode:   "USD",
									Rate:             0.013, // GBp -> USD (GBP -> USD scaled by 10^-2)
								},
								QuotePrice: c.QuotePrice{
									Price: 5000.0, // GBp
								},
							},
						},
					}
					assets, _ := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(1))
					Expect(assets[0].Position.Value).To(Equal(50000.0)) // 10 * 5000 GBp, left in quote currency
					Expect(assets[0].Position.Weight).To(Equal(100.0))  // not 1/0.013 * 100 ~= 7692%
				})
			})

			When("minor and major currency assets are mixed with summary-only conversion", func() {
				It("should compute weights in the common summary currency", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency:                   "GBP",
							CurrencyConvertSummaryOnly: true,
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: c.AssetGroup{
							ConfigAssetGroup: c.ConfigAssetGroup{
								Lots: []c.Lot{
									{
										Symbol:   "JEGP.L",
										Quantity: 10,
										UnitCost: 4000.0, // GBp
									},
									{
										Symbol:   "VUAG.L",
										Quantity: 10,
										UnitCost: 40.0, // GBP
									},
								},
							},
						},
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "JEGP.L",
								Currency: c.Currency{
									FromCurrencyCode: "GBp",
									ToCurrencyCode:   "GBP",
									Rate:             0.01, // GBp -> GBP
								},
								QuotePrice: c.QuotePrice{
									Price: 5000.0, // GBp -> 500 GBP worth of value at qty 10
								},
							},
							{
								Symbol: "VUAG.L",
								Currency: c.Currency{
									FromCurrencyCode: "GBP",
									ToCurrencyCode:   "GBP",
									Rate:             1.0, // GBP -> GBP
								},
								QuotePrice: c.QuotePrice{
									Price: 50.0, // GBP -> 500 GBP worth of value at qty 10
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(2))
					// Per-position values stay in their quote currency
					Expect(assets[0].Position.Value).To(Equal(50000.0)) // GBp
					Expect(assets[1].Position.Value).To(Equal(500.0))   // GBP
					// Both contribute 500 GBP to the summary, so weights are equal
					Expect(assets[0].Position.Weight).To(Equal(50.0))
					Expect(assets[1].Position.Weight).To(Equal(50.0))
					Expect(summary.Value).To(Equal(1000.0)) // 500 GBP + 500 GBP
				})
			})

			When("a minor currency asset is fully converted to the major unit", func() {
				It("should convert value, cost, and avg cost to the major unit and weight at 100%", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency: "GBP",
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: c.AssetGroup{
							ConfigAssetGroup: c.ConfigAssetGroup{
								Lots: []c.Lot{
									{
										Symbol:    "JEGP.L",
										Quantity:  10,
										UnitCost:  4000.0, // GBp (quote currency, converted automatically)
										FixedCost: 100.0,  // GBp
									},
								},
							},
						},
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "JEGP.L",
								Currency: c.Currency{
									FromCurrencyCode: "GBp",
									ToCurrencyCode:   "GBP",
									Rate:             0.01, // GBp -> GBP
								},
								QuotePrice: c.QuotePrice{
									Price: 5000.0, // GBp
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(1))
					Expect(assets[0].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[0].QuotePrice.Price).To(Equal(50.0))  // 5000 GBp -> 50 GBP
					Expect(assets[0].Position.Value).To(Equal(500.0))   // 10 * 5000 GBp -> 500 GBP
					Expect(assets[0].Position.Cost).To(Equal(401.0))    // (4000*10 + 100) GBp -> 401 GBP
					Expect(assets[0].Position.UnitCost).To(Equal(40.1)) // GBP
					Expect(assets[0].Position.Weight).To(Equal(100.0))
					Expect(summary.Value).To(Equal(500.0)) // already in GBP
				})
			})

			When("a minor currency asset is converted but unit cost conversion is disabled", func() {
				It("should convert value to the major unit while leaving the cost basis as entered", func() {
					inputCtx := c.Context{
						Config: c.Config{
							Currency:                          "GBP",
							CurrencyDisableUnitCostConversion: true,
						},
						Reference: c.Reference{},
					}
					assetGroupQuote := c.AssetGroupQuote{
						AssetGroup: c.AssetGroup{
							ConfigAssetGroup: c.ConfigAssetGroup{
								Lots: []c.Lot{
									{
										Symbol:    "JEGP.L",
										Quantity:  10,
										UnitCost:  40.0,  // GBP (major unit, entered directly)
										FixedCost: 11.95, // GBP
									},
								},
							},
						},
						AssetQuotes: []c.AssetQuote{
							{
								Symbol: "JEGP.L",
								Currency: c.Currency{
									FromCurrencyCode: "GBp",
									ToCurrencyCode:   "GBP",
									Rate:             0.01, // GBp -> GBP
								},
								QuotePrice: c.QuotePrice{
									Price: 5000.0, // GBp
								},
							},
						},
					}
					assets, summary := GetAssets(inputCtx, assetGroupQuote)

					Expect(assets).To(HaveLen(1))
					Expect(assets[0].QuotePrice.Price).To(Equal(50.0)) // 5000 GBp -> 50 GBP
					Expect(assets[0].Position.Value).To(Equal(500.0))  // 10 * 5000 GBp -> 500 GBP
					Expect(assets[0].Position.Cost).To(Equal(411.95))  // (40*10 + 11.95) GBP, not converted
					Expect(assets[0].Position.Weight).To(Equal(100.0))
					Expect(summary.Value).To(Equal(500.0))
				})
			})
		})
	})
})

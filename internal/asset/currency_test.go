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
					Expect(assets[0].Position.Value).To(Equal(1000.0))   // No conversion applied
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[1].Position.Value).To(Equal(1000.0))   // No conversion applied
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
					Expect(assets[0].Position.Value).To(Equal(1000.0))   // No conversion applied
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[1].Position.Value).To(Equal(1000.0))   // No conversion applied
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
					Expect(assets[0].Position.Cost).To(Equal(1000.0))    // Not converted
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("EUR"))
					Expect(assets[1].QuotePrice.Price).To(Equal(50.0)) // Converted
					Expect(assets[1].Position.Cost).To(Equal(500.0))    // Not converted
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
					Expect(assets[0].Position.Value).To(Equal(1000.0))   // No conversion applied
					Expect(assets[1].Currency.FromCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].Currency.ToCurrencyCode).To(Equal("GBP"))
					Expect(assets[1].QuotePrice.Price).To(Equal(100.0)) // No conversion applied
					Expect(assets[1].Position.Value).To(Equal(1000.0))   // No conversion applied
					Expect(summary.Value).To(Equal(1750.0))             // Converted
					Expect(summary.Cost).To(Equal(1500.0))              // Converted
				})
			})

		})
	})
})

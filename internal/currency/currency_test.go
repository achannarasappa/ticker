package currency_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/currency"
)

var _ = Describe("Currency", func() {

	Describe("GetCurrencyRateFromContext", func() {
		It("should return default currency information when a rate is not found in reference data", func() {
			inputCtx := c.Context{
				Reference: c.Reference{
					CurrencyRates: c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         4,
						},
						"GBP": c.CurrencyRate{
							FromCurrency: "GBP",
							ToCurrency:   "EUR",
							Rate:         2,
						},
					},
				},
			}
			outputCurrencyRateByUse := GetCurrencyRateFromContext(inputCtx, "EUR")
			Expect(outputCurrencyRateByUse.QuotePrice).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.PositionCost).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.SummaryValue).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.SummaryCost).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.ToCurrencyCode).To(Equal("EUR"))
		})
	})

	When("there is a matching currency in reference data", func() {
		It("should return rate to convert", func() {
			inputCtx := c.Context{
				Config: c.Config{
					Currency: "EUR",
				},
				Reference: c.Reference{
					CurrencyRates: c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         1.25,
						},
						"GBP": c.CurrencyRate{
							FromCurrency: "GBP",
							ToCurrency:   "EUR",
							Rate:         2,
						},
					},
				},
			}
			outputCurrencyRateByUse := GetCurrencyRateFromContext(inputCtx, "USD")
			Expect(outputCurrencyRateByUse.QuotePrice).To(Equal(1.25))
			Expect(outputCurrencyRateByUse.PositionCost).To(Equal(1.25))
			Expect(outputCurrencyRateByUse.SummaryValue).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.SummaryCost).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.ToCurrencyCode).To(Equal("EUR"))
		})
	})

	When("the currency is not set", func() {
		It("should return the conversion rate to convert only the summary line", func() {
			inputCtx := c.Context{
				Config: c.Config{
					Currency: "",
				},
				Reference: c.Reference{
					CurrencyRates: c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         1.25,
						},
						"GBP": c.CurrencyRate{
							FromCurrency: "GBP",
							ToCurrency:   "EUR",
							Rate:         2,
						},
					},
				},
			}
			outputCurrencyRateByUse := GetCurrencyRateFromContext(inputCtx, "USD")
			Expect(outputCurrencyRateByUse.QuotePrice).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.PositionCost).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.SummaryValue).To(Equal(1.25))
			Expect(outputCurrencyRateByUse.SummaryCost).To(Equal(1.25))
			Expect(outputCurrencyRateByUse.ToCurrencyCode).To(Equal("EUR"))
		})
	})

	When("the option to convert only the summary values is set", func() {
		It("should return summary conversion rate", func() {
			inputCtx := c.Context{
				Config: c.Config{
					Currency:                   "EUR",
					CurrencyConvertSummaryOnly: true,
				},
				Reference: c.Reference{
					CurrencyRates: c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         1.25,
						},
						"GBP": c.CurrencyRate{
							FromCurrency: "GBP",
							ToCurrency:   "EUR",
							Rate:         2,
						},
					},
				},
			}
			outputCurrencyRateByUse := GetCurrencyRateFromContext(inputCtx, "USD")
			Expect(outputCurrencyRateByUse.QuotePrice).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.PositionCost).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.SummaryValue).To(Equal(1.25))
			Expect(outputCurrencyRateByUse.SummaryCost).To(Equal(1.25))
			Expect(outputCurrencyRateByUse.ToCurrencyCode).To(Equal("USD"))
		})
	})

	When("the option to not convert unit cost is set", func() {
		It("should not convert cost", func() {
			inputCtx := c.Context{
				Config: c.Config{
					Currency:                          "EUR",
					CurrencyDisableUnitCostConversion: true,
				},
				Reference: c.Reference{
					CurrencyRates: c.CurrencyRates{
						"USD": c.CurrencyRate{
							FromCurrency: "USD",
							ToCurrency:   "EUR",
							Rate:         1.25,
						},
						"GBP": c.CurrencyRate{
							FromCurrency: "GBP",
							ToCurrency:   "EUR",
							Rate:         2,
						},
					},
				},
			}
			outputCurrencyRateByUse := GetCurrencyRateFromContext(inputCtx, "USD")
			Expect(outputCurrencyRateByUse.SummaryCost).To(Equal(1.0))
			Expect(outputCurrencyRateByUse.PositionCost).To(Equal(1.0))
		})
	})
})

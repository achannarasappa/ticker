package currency_test

import (
	"net/http"

	"github.com/jarcoal/httpmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/currency"
	. "github.com/achannarasappa/ticker/test/http"
)

var _ = Describe("Currency", func() {
	Describe("GetCurrencyRates", func() {
		It("should get the currency exchange rate", func() {

			MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
			MockResponse(ResponseParameters{Symbol: "EURUSD=X", Currency: "USD", Price: 1.2})
			output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "USD")
			Expect(err).ToNot(HaveOccurred())
			Expect(output).To(Equal(c.CurrencyRates{
				"EUR": c.CurrencyRate{
					FromCurrency: "EUR",
					ToCurrency:   "USD",
					Rate:         1.2,
				},
			}))

		})

		When("target currency is not set", func() {
			It("defaults to USD", func() {

				MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
				MockResponse(ResponseParameters{Symbol: "EURUSD=X", Currency: "USD", Price: 1.2})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{
					"EUR": c.CurrencyRate{
						FromCurrency: "EUR",
						ToCurrency:   "USD",
						Rate:         1.2,
					},
				}))

			})
		})

		When("target currency is the same as all watchlist currencies", func() {
			It("returns an empty currency exchange rate list", func() {

				MockResponse(ResponseParameters{Symbol: "VOW3.DE", Currency: "EUR", Price: 0.0})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "EUR")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})

		When("the request to get the currencies of each symbol fails", func() {
			It("returns error", func() {

				responseText := `{
					"quoteResponse": {
						"result": [
							{
								"regularMarketPrice": 1.2,
								"currency": "EUR",
								"symbol": "EURUSD=X"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=EURUSD=X"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseText)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "EUR")
				Expect(err).To(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})

		When("the request to get the currencies of each symbol does not include a currency", func() {
			It("should return an empty currency rate", func() {

				responseText := `{
					"quoteResponse": {
						"result": [
							{
								"regularMarketPrice": 160.0,
								"symbol": "VOW3.DE"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=VOW3.DE"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseText)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "EUR")
				Expect(err).ToNot(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})

		When("the request to the exchange rate fails", func() {
			It("returns error", func() {

				responseText := `{
					"quoteResponse": {
						"result": [
							{
								"regularMarketPrice": 160.0,
								"currency": "EUR",
								"symbol": "VOW3.DE"
							}
						],
						"error": null
					}
				}`
				responseUrl := "https://query1.finance.yahoo.com/v7/finance/quote?lang=en-US&region=US&corsDomain=finance.yahoo.com&fields=regularMarketPrice,currency&symbols=VOW3.DE"
				httpmock.RegisterResponder("GET", responseUrl, func(req *http.Request) (*http.Response, error) {
					resp := httpmock.NewStringResponse(200, responseText)
					resp.Header.Set("Content-Type", "application/json")
					return resp, nil
				})
				output, err := GetCurrencyRates(*client, []string{"VOW3.DE"}, "USD")
				Expect(err).To(HaveOccurred())
				Expect(output).To(Equal(c.CurrencyRates{}))

			})
		})
	})

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
			Expect(outputCurrencyRateByUse.PositionValue).To(Equal(1.0))
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
			Expect(outputCurrencyRateByUse.PositionValue).To(Equal(1.25))
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
			Expect(outputCurrencyRateByUse.PositionValue).To(Equal(1.0))
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
			Expect(outputCurrencyRateByUse.PositionValue).To(Equal(1.0))
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
		})
	})
})

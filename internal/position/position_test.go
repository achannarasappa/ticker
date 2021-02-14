package position_test

import (
	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position", func() {

	Describe("GetLots", func() {
		It("should return a map of aggregated lots", func() {
			input := []c.Lot{
				{Symbol: "ABNB", UnitCost: 146.00, Quantity: 35},
				{Symbol: "ARKW", UnitCost: 152.25, Quantity: 20},
				{Symbol: "ARKW", UnitCost: 152.25, Quantity: 20},
			}
			output := GetLots(input)
			expected := map[string]AggregatedLot{
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35},
				"ARKW": {Symbol: "ARKW", Cost: 6090, Quantity: 40},
			}
			Expect(output).To(Equal(expected))
		})

		When("lots are not set", func() {
			It("returns en empty aggregated lot", func() {
				var input []c.Lot = nil
				output := GetLots(input)
				expected := map[string]AggregatedLot{}
				Expect(output).To(Equal(expected))
			})
		})
	})

	Describe("GetSymbols", func() {
		It("should return a slice of symbols", func() {
			inputAggregatedLots := map[string]AggregatedLot{
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35},
				"ARKW": {Symbol: "ARKW", Cost: 6090, Quantity: 40},
			}
			inputSymbols := []string{"GOOG", "ARKW"}
			output := GetSymbols(inputSymbols, inputAggregatedLots)
			expected := []string{
				"ABNB",
				"ARKW",
				"GOOG",
			}
			Expect(output).To(ContainElements(expected))
		})
	})

	Describe("GetPositions", func() {
		It("should return a map of positions", func() {
			inputAggregatedLots := map[string]AggregatedLot{
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35},
				"ARKW": {Symbol: "ARKW", Cost: 4000, Quantity: 40},
			}
			inputQuotes := []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "ARKW",
						RegularMarketPreviousClose: 100,
					},
					Price:  200.0,
					Change: 50.0,
				},
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "RBLX",
						RegularMarketPreviousClose: 50,
					},
					Price:  50.0,
					Change: 0.0,
				},
			}
			output := GetPositions(inputAggregatedLots)(inputQuotes)
			expected := map[string]Position{
				"ARKW": {
					AggregatedLot: AggregatedLot{
						Symbol:   "ARKW",
						Cost:     4000,
						Quantity: 40,
					},
					Value:              8000,
					DayChange:          2000,
					DayChangePercent:   50,
					TotalChange:        4000,
					TotalChangePercent: 100,
				},
			}
			Expect(output).To(Equal(expected))
		})
	})

	Describe("GetPositionSummary", func() {
		It("should return a summary of positions", func() {
			inputPositions := map[string]Position{
				"ARKW": {
					AggregatedLot: AggregatedLot{
						Symbol:   "ARKW",
						Cost:     1000,
						Quantity: 10,
					},
					Value:     2000,
					DayChange: 200,
				},
				"RBLX": {
					AggregatedLot: AggregatedLot{
						Symbol:   "RBLX",
						Cost:     1000,
						Quantity: 10,
					},
					Value:     2000,
					DayChange: 200,
				},
				"ANI": {
					AggregatedLot: AggregatedLot{
						Symbol:   "ANI",
						Cost:     1000,
						Quantity: 10,
					},
					Value:     2000,
					DayChange: 200,
					Currency:  "EUR",
				},
			}
			inputCtx := c.Context{
				Reference: c.Reference{
					CurrencyRates: c.CurrencyRates{
						"EUR": c.CurrencyRate{
							FromCurrency: "EUR",
							ToCurrency:   "USD",
							Rate:         1.5,
						},
					},
				},
			}
			output := GetPositionSummary(inputCtx, inputPositions)
			expected := PositionSummary{
				Value:            7000,
				Cost:             3500,
				Change:           3500,
				DayChange:        700,
				ChangePercent:    200,
				DayChangePercent: 10,
			}

			Expect(output).To(Equal(expected))
		})
	})
})

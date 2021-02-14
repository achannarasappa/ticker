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
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35, OrderIndex: 0},
				"ARKW": {Symbol: "ARKW", Cost: 6090, Quantity: 40, OrderIndex: 1},
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
				"ARKW": {Symbol: "ARKW", Cost: 1000, Quantity: 10},
				"ANI":  {Symbol: "ANI", Cost: 1000, Quantity: 10},
				"TW":   {Symbol: "TW", Cost: 2000, Quantity: 20},
			}
			inputQuotes := []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "ARKW",
						RegularMarketPreviousClose: 100,
					},
					Price:  120.0,
					Change: 20.0,
				},
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "TW",
						RegularMarketPreviousClose: 200,
					},
					Price:  200.0,
					Change: 20.0,
				},
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "ANI",
						RegularMarketPreviousClose: 25,
					},
					Price:  25.0,
					Change: 2.5,
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
			inputCtx := c.Context{}
			outputPositions, outputPositionSummary := GetPositions(inputCtx, inputAggregatedLots)(inputQuotes)
			expectedPosition := map[string]Position{
				"ARKW": {
					AggregatedLot: AggregatedLot{
						Symbol:   "ARKW",
						Cost:     1000,
						Quantity: 10,
					},
					Value:              1200,
					DayChange:          200,
					TotalChange:        200,
					TotalChangePercent: 20,
				},
				"TW": {
					AggregatedLot: AggregatedLot{
						Symbol:     "TW",
						Cost:       2000,
						Quantity:   20,
						OrderIndex: 0,
					},
					Value:              4000,
					DayChange:          400,
					DayChangePercent:   0,
					TotalChange:        2000,
					TotalChangePercent: 100,
					Currency:           "",
					CurrencyConverted:  "",
					Weight:             0,
				},
				"ANI": {
					AggregatedLot: AggregatedLot{
						Symbol:   "ANI",
						Cost:     1000,
						Quantity: 10,
					},
					Value:              250,
					DayChange:          25,
					TotalChange:        -750,
					TotalChangePercent: -75,
				},
			}
			expectedPositionSummary := PositionSummary{
				Value:            5450,
				Cost:             4000,
				Change:           1450,
				DayChange:        4000,
				ChangePercent:    136.25,
				DayChangePercent: 11.46788990825688,
			}
			Expect(outputPositions).To(Equal(expectedPosition))
			Expect(outputPositionSummary).To(Equal(expectedPositionSummary))
		})
	})
})

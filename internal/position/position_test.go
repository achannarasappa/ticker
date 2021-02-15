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
			expectedPositionSummary := PositionSummary{
				Value:            5450,
				Cost:             4000,
				Change:           1450,
				DayChange:        625,
				ChangePercent:    136.25,
				DayChangePercent: 11.46788990825688,
			}
			Expect(outputPositions["ARKW"]).To(Equal(Position{
				AggregatedLot: AggregatedLot{
					Symbol:   "ARKW",
					Cost:     1000,
					Quantity: 10,
				},
				Value:              1200,
				Cost:               1000,
				DayChange:          200,
				TotalChange:        200,
				TotalChangePercent: 20,
				AverageCost:        100,
				Weight:             22.018348623853214,
			}))
			Expect(outputPositions["TW"]).To(Equal(Position{
				AggregatedLot: AggregatedLot{
					Symbol:     "TW",
					Cost:       2000,
					Quantity:   20,
					OrderIndex: 0,
				},
				Value:              4000,
				Cost:               2000,
				DayChange:          400,
				DayChangePercent:   0,
				TotalChange:        2000,
				TotalChangePercent: 100,
				AverageCost:        100,
				Weight:             73.39449541284404,
			}))
			Expect(outputPositions["ANI"]).To(Equal(Position{
				AggregatedLot: AggregatedLot{
					Symbol:   "ANI",
					Cost:     1000,
					Quantity: 10,
				},
				Value:              250,
				Cost:               1000,
				DayChange:          25,
				TotalChange:        -750,
				TotalChangePercent: -75,
				AverageCost:        100,
				Weight:             4.587155963302752,
			}))
			Expect(outputPositionSummary).To(Equal(expectedPositionSummary))
		})

		When("no aggregated lots are set", func() {
			It("should return an empty positions and position summary", func() {
				inputAggregatedLots := map[string]AggregatedLot{}
				inputQuotes := []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:                     "ARKW",
							RegularMarketPreviousClose: 100,
						},
						Price:  120.0,
						Change: 20.0,
					},
				}
				inputCtx := c.Context{}
				outputPositions, outputPositionSummary := GetPositions(inputCtx, inputAggregatedLots)(inputQuotes)
				expectedPositionSummary := PositionSummary{}
				Expect(outputPositions).To(Equal(map[string]Position{}))
				Expect(outputPositionSummary).To(Equal(expectedPositionSummary))
			})
		})
	})
})

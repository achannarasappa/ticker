package position_test

import (
	"strings"

	. "ticker-tape/internal/position"
	. "ticker-tape/internal/quote"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Position", func() {

	var (
		lotFixture = `
  - symbol: "ABNB"
    quantity: 35.0
    unit_cost: 146.00
  - symbol: "ARKW"
    quantity: 20.0
    unit_cost: 152.25
  - symbol: "ARKW"
    quantity: 20.0
    unit_cost: 152.25
`
	)

	Describe("GetLots", func() {
		It("should return a map of aggregated lots", func() {
			input := strings.NewReader(lotFixture)
			output := GetLots(input)
			expected := map[string]AggregatedLot{
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35},
				"ARKW": {Symbol: "ARKW", Cost: 6090, Quantity: 40},
			}
			Expect(output).To(Equal(expected))
		})
	})

	Describe("GetSymbols", func() {
		It("should return a slice of symbols", func() {
			input := map[string]AggregatedLot{
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35},
				"ARKW": {Symbol: "ARKW", Cost: 6090, Quantity: 40},
			}
			output := GetSymbols(input)
			expected := []string{
				"ABNB",
				"ARKW",
			}
			Expect(output).To(Equal(expected))
		})
	})

	Describe("GetPositions", func() {
		It("should return a map of positions", func() {
			inputAggregatedLots := map[string]AggregatedLot{
				"ABNB": {Symbol: "ABNB", Cost: 5110, Quantity: 35},
				"ARKW": {Symbol: "ARKW", Cost: 6090, Quantity: 40},
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
						Cost:     6090,
						Quantity: 40,
					},
					Value:            8000,
					DayChange:        2000,
					DayChangePercent: 50,
				},
			}
			Expect(output).To(Equal(expected))
		})
	})

})

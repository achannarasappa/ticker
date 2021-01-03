package position

import (
	"github.com/novalagung/gubrak/v2"
)

type Lot struct {
	Symbol   string
	UnitCost float64
	Quantity float64
}

type AggregatedLot struct {
	Symbol   string
	Cost     float64
	Quantity float64
}

func aggregateLotsBySymbol(lots []Lot) map[string]AggregatedLot {

	aggregatedLots := gubrak.
		From(lots).
		Reduce(func(acc map[string]AggregatedLot, lot Lot) map[string]AggregatedLot {

			aggregatedLot, ok := acc[lot.Symbol]
			if !ok {
				acc[lot.Symbol] = AggregatedLot{
					Symbol:   lot.Symbol,
					Cost:     lot.UnitCost * lot.Quantity,
					Quantity: lot.Quantity,
				}
				return acc
			}

			aggregatedLot.Quantity = aggregatedLot.Quantity + lot.Quantity
			aggregatedLot.Cost = aggregatedLot.Cost + (lot.Quantity * lot.UnitCost)

			acc[lot.Symbol] = aggregatedLot

			return acc

		}, make(map[string]AggregatedLot)).
		Result()

	return (aggregatedLots).(map[string]AggregatedLot)

}

func GetSymbols() []string {

	symbols := gubrak.
		From(lots).
		Reduce(func(acc []string, lot Lot) []string {
			return append(acc, lot.Symbol)
		}, make([]string, 0)).
		Result()

	return (symbols).([]string)

}

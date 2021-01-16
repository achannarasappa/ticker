package position

import (
	"io"
	"log"
	. "ticker-tape/internal/quote"

	"github.com/novalagung/gubrak/v2"
	"gopkg.in/yaml.v2"
)

type Position struct {
	AggregatedLot
	Value            float64
	DayChange        float64
	DayChangePercent float64
}

type Lot struct {
	Symbol   string  `yaml:"symbol"`
	UnitCost float64 `yaml:"unit_cost"`
	Quantity float64 `yaml:"quantity"`
}

type AggregatedLot struct {
	Symbol   string
	Cost     float64
	Quantity float64
}

func GetLots(handle io.Reader) map[string]AggregatedLot {

	lots := []Lot{}
	decoderLots := yaml.NewDecoder(handle)
	err := decoderLots.Decode(&lots)

	if err != nil {
		log.Fatalf("error: %v", err)
	}

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

func GetSymbols(symbols []string, aggregatedLots map[string]AggregatedLot) []string {

	symbolsFromAggregatedLots := make([]string, 0)
	for k := range aggregatedLots {
		symbolsFromAggregatedLots = append(symbolsFromAggregatedLots, k)
	}

	return (gubrak.From(symbolsFromAggregatedLots).
		Concat(symbols).
		Uniq().
		Result()).([]string)

}

func GetPositions(aggregatedLots map[string]AggregatedLot) func([]Quote) map[string]Position {
	return func(quotes []Quote) map[string]Position {

		positions := gubrak.
			From(quotes).
			Reduce(func(acc []Position, quote Quote) []Position {
				if _, ok := aggregatedLots[quote.Symbol]; ok {
					dayChange := quote.Change * aggregatedLots[quote.Symbol].Quantity
					valuePreviousClose := quote.RegularMarketPreviousClose * aggregatedLots[quote.Symbol].Quantity
					return append(acc, Position{
						AggregatedLot:    aggregatedLots[quote.Symbol],
						Value:            quote.Price * aggregatedLots[quote.Symbol].Quantity,
						DayChange:        dayChange,
						DayChangePercent: (dayChange / valuePreviousClose) * 100,
					})
				}
				return acc
			}, make([]Position, 0)).
			KeyBy(func(position Position) string {
				return position.Symbol
			}).
			Result()

		return (positions).(map[string]Position)
	}
}

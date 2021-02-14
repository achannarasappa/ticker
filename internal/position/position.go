package position

import (
	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/currency"
	. "github.com/achannarasappa/ticker/internal/quote"

	"github.com/novalagung/gubrak/v2"
)

type Position struct {
	AggregatedLot
	Value              float64
	DayChange          float64
	DayChangePercent   float64
	TotalChange        float64
	TotalChangePercent float64
	Currency           string
	CurrencyConverted  string
}

type PositionSummary struct {
	Value            float64
	Cost             float64
	Change           float64
	DayChange        float64
	ChangePercent    float64
	DayChangePercent float64
}

type AggregatedLot struct {
	Symbol     string
	Cost       float64
	Quantity   float64
	OrderIndex int
}

func GetLots(lots []c.Lot) map[string]AggregatedLot {

	if lots == nil {
		return map[string]AggregatedLot{}
	}

	aggregatedLots := gubrak.
		From(lots).
		Reduce(func(acc map[string]AggregatedLot, lot c.Lot, i int) map[string]AggregatedLot {

			aggregatedLot, ok := acc[lot.Symbol]
			if !ok {
				acc[lot.Symbol] = AggregatedLot{
					Symbol:     lot.Symbol,
					Cost:       lot.UnitCost * lot.Quantity,
					Quantity:   lot.Quantity,
					OrderIndex: i,
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

func GetPositions(ctx c.Context, aggregatedLots map[string]AggregatedLot) func([]Quote) map[string]Position {
	return func(quotes []Quote) map[string]Position {

		positions := gubrak.
			From(quotes).
			Reduce(func(acc []Position, quote Quote) []Position {
				if aggLot, ok := aggregatedLots[quote.Symbol]; ok {
					currencyRate, currencyCode := currency.GetCurrencyRateFromContext(ctx, quote.Currency)
					totalChange := (quote.Price * aggLot.Quantity) - (aggLot.Cost * currencyRate)
					return append(acc, Position{
						AggregatedLot:      aggLot,
						Value:              quote.Price * aggLot.Quantity,
						DayChange:          quote.Change * aggLot.Quantity,
						DayChangePercent:   quote.ChangePercent,
						TotalChange:        totalChange,
						TotalChangePercent: (totalChange / (aggLot.Cost * currencyRate)) * 100,
						Currency:           quote.Currency,
						CurrencyConverted:  currencyCode,
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

func GetPositionSummary(ctx c.Context, positions map[string]Position) PositionSummary {

	positionValueCost := gubrak.From(positions).
		Reduce(func(acc PositionSummary, position Position, key string) PositionSummary {
			currencyRate := 1.0
			if ctx.Config.Currency == "" {
				currencyRate, _ = currency.GetCurrencyRateFromContext(ctx, position.Currency)
			}
			acc.Value += (position.Value * currencyRate)
			acc.Cost += (position.Cost * currencyRate)
			acc.DayChange += (position.DayChange * currencyRate)
			return acc
		}, PositionSummary{}).
		Result()

	positionSummary := (positionValueCost).(PositionSummary)

	positionSummary.Change = positionSummary.Value - positionSummary.Cost
	positionSummary.ChangePercent = (positionSummary.Value / positionSummary.Cost) * 100
	positionSummary.DayChangePercent = (positionSummary.DayChange / positionSummary.Value) * 100

	return positionSummary

}

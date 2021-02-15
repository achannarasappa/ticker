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
	Cost               float64
	DayChange          float64
	DayChangePercent   float64
	TotalChange        float64
	TotalChangePercent float64
	Currency           string
	CurrencyConverted  string
	AverageCost        float64
	Weight             float64
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

type positionSummaryBase struct {
	value     float64
	cost      float64
	dayChange float64
}

type positionAcc struct {
	positionSummaryBase positionSummaryBase
	positions           []Position
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

func GetPositions(ctx c.Context, aggregatedLots map[string]AggregatedLot) func([]Quote) (map[string]Position, PositionSummary) {
	return func(quotes []Quote) (map[string]Position, PositionSummary) {

		if len(aggregatedLots) <= 0 {
			return map[string]Position{}, PositionSummary{}
		}

		positionsReduced := (gubrak.
			From(quotes).
			Reduce(func(acc positionAcc, quote Quote) positionAcc {

				if aggLot, ok := aggregatedLots[quote.Symbol]; ok {

					currencyRate, currencyRateDefault, currencyCode := currency.GetCurrencyRateFromContext(ctx, quote.Currency)

					cost := aggLot.Cost * currencyRate
					value := quote.Price * aggLot.Quantity
					totalChange := value - cost
					totalChangePercant := (totalChange / cost) * 100

					position := Position{
						AggregatedLot:      aggLot,
						Value:              value,
						Cost:               cost,
						DayChange:          quote.Change * aggLot.Quantity,
						DayChangePercent:   quote.ChangePercent,
						TotalChange:        totalChange,
						TotalChangePercent: totalChangePercant,
						AverageCost:        cost / aggLot.Quantity,
						Currency:           quote.Currency,
						CurrencyConverted:  currencyCode,
					}

					acc.positions = append(acc.positions, position)
					acc.positionSummaryBase = getPositionSummaryBase(position, acc.positionSummaryBase, currencyRateDefault)

				}

				return acc

			}, positionAcc{}).
			Result()).(positionAcc)

		positionSummary := PositionSummary{
			Value:            positionsReduced.positionSummaryBase.value,
			Cost:             positionsReduced.positionSummaryBase.cost,
			Change:           positionsReduced.positionSummaryBase.value - positionsReduced.positionSummaryBase.cost,
			DayChange:        positionsReduced.positionSummaryBase.dayChange,
			ChangePercent:    (positionsReduced.positionSummaryBase.value / positionsReduced.positionSummaryBase.cost) * 100,
			DayChangePercent: (positionsReduced.positionSummaryBase.dayChange / positionsReduced.positionSummaryBase.value) * 100,
		}

		positions := gubrak.From(positionsReduced.positions).
			Map(func(v Position) Position {
				v.Weight = (v.Value / positionSummary.Value) * 100
				return v
			}).
			KeyBy(func(position Position) string {
				return position.Symbol
			}).
			Result()

		return (positions).(map[string]Position), positionSummary
	}
}

func getPositionSummaryBase(position Position, acc positionSummaryBase, currencyRateDefault float64) positionSummaryBase {
	acc.value += (position.Value * currencyRateDefault)
	acc.cost += (position.Cost * currencyRateDefault)
	acc.dayChange += (position.DayChange * currencyRateDefault)
	return acc
}

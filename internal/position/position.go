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

func GetSymbols(config c.Config, aggregatedLots map[string]AggregatedLot) []string {

	symbols := make(map[string]bool)
	symbolsUnique := make([]string, 0)

	for _, v := range config.Watchlist {
		if !symbols[v] {
			symbols[v] = true
			symbolsUnique = append(symbolsUnique, v)
		}
	}

	if config.ShowHoldings {
		for k := range aggregatedLots {
			if !symbols[k] {
				symbols[k] = true
				symbolsUnique = append(symbolsUnique, k)
			}
		}
	}

	return symbolsUnique

}

func GetPositions(ctx c.Context, aggregatedLots map[string]AggregatedLot) func([]Quote) (map[string]Position, PositionSummary) {
	return func(quotes []Quote) (map[string]Position, PositionSummary) {

		if len(aggregatedLots) <= 0 {
			return map[string]Position{}, PositionSummary{}
		}

		positionsReduced := getPositionsReduced(ctx, aggregatedLots, quotes)

		change := positionsReduced.positionSummaryBase.value - positionsReduced.positionSummaryBase.cost

		positionSummary := PositionSummary{
			Value:            positionsReduced.positionSummaryBase.value,
			Cost:             positionsReduced.positionSummaryBase.cost,
			Change:           positionsReduced.positionSummaryBase.value - positionsReduced.positionSummaryBase.cost,
			DayChange:        positionsReduced.positionSummaryBase.dayChange,
			ChangePercent:    (change / positionsReduced.positionSummaryBase.cost) * 100,
			DayChangePercent: (positionsReduced.positionSummaryBase.dayChange / positionsReduced.positionSummaryBase.value) * 100,
		}

		positions := getPositionMapFromPositionsReduced(positionsReduced.positions, positionSummary.Value)

		return positions, positionSummary
	}
}

func getPositionsReduced(ctx c.Context, aggregatedLots map[string]AggregatedLot, quotes []Quote) positionAcc {

	acc := positionAcc{}
	for _, quote := range quotes {
		if aggregatedLot, ok := aggregatedLots[quote.Symbol]; ok {
			currencyRate, currencyRateDefault, currencyCode := currency.GetCurrencyRateFromContext(ctx, quote.Currency)

			cost := aggregatedLot.Cost * currencyRate
			value := quote.Price * aggregatedLot.Quantity
			totalChange := value - cost
			totalChangePercant := (totalChange / cost) * 100

			position := Position{
				AggregatedLot:      aggregatedLot,
				Value:              value,
				Cost:               cost,
				DayChange:          quote.Change * aggregatedLot.Quantity,
				DayChangePercent:   quote.ChangePercent,
				TotalChange:        totalChange,
				TotalChangePercent: totalChangePercant,
				AverageCost:        cost / aggregatedLot.Quantity,
				Currency:           quote.Currency,
				CurrencyConverted:  currencyCode,
			}

			acc.positions = append(acc.positions, position)
			acc.positionSummaryBase = getPositionSummaryBase(position, acc.positionSummaryBase, currencyRateDefault)
		}
	}

	return acc
}

func getPositionMapFromPositionsReduced(p []Position, totalValue float64) map[string]Position {
	positions := map[string]Position{}
	for _, v := range p {
		v.Weight = (v.Value / totalValue) * 100
		positions[v.Symbol] = v
	}
	return positions
}

func getPositionSummaryBase(position Position, acc positionSummaryBase, currencyRateDefault float64) positionSummaryBase {
	acc.value += (position.Value * currencyRateDefault)
	acc.cost += (position.Cost * currencyRateDefault)
	acc.dayChange += (position.DayChange * currencyRateDefault)
	return acc
}

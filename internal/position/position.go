package position

import (
	c "github.com/achannarasappa/ticker/internal/common"
	"github.com/achannarasappa/ticker/internal/currency"
	q "github.com/achannarasappa/ticker/internal/quote"
)

// Position represents a position taken with a security at a point in time
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

// PositionSummary represents a summary of all positions at a point in time
type PositionSummary struct {
	Value            float64
	Cost             float64
	Change           float64
	DayChange        float64
	ChangePercent    float64
	DayChangePercent float64
}

// AggregatedLot represents a summation of the costs and quantities of a single type of security
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

// GetLots aggregates costs basis lots
func GetLots(lots []c.Lot) map[string]AggregatedLot {

	if lots == nil {
		return map[string]AggregatedLot{}
	}

	aggregatedLots := map[string]AggregatedLot{}

	for i, lot := range lots {

		aggregatedLot, ok := aggregatedLots[lot.Symbol]

		if !ok {

			aggregatedLots[lot.Symbol] = AggregatedLot{
				Symbol:     lot.Symbol,
				Cost:       (lot.UnitCost * lot.Quantity) + lot.FixedCost,
				Quantity:   lot.Quantity,
				OrderIndex: i,
			}

		} else {

			aggregatedLot.Quantity = aggregatedLot.Quantity + lot.Quantity
			aggregatedLot.Cost = aggregatedLot.Cost + (lot.Quantity * lot.UnitCost)

			aggregatedLots[lot.Symbol] = aggregatedLot

		}

	}

	return aggregatedLots
}

// GetSymbols generates a unique list of symbols from the watchlist and cost basis lots
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

// GetPositions calculates the positions for each symbol and a summary of all positions
func GetPositions(ctx c.Context, aggregatedLots map[string]AggregatedLot) func([]q.Quote) (map[string]Position, PositionSummary) {
	return func(quotes []q.Quote) (map[string]Position, PositionSummary) {

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

func getPositionsReduced(ctx c.Context, aggregatedLots map[string]AggregatedLot, quotes []q.Quote) positionAcc {

	acc := positionAcc{}
	for _, quote := range quotes {
		if aggregatedLot, ok := aggregatedLots[quote.Symbol]; ok {
			currencyRateByUse := currency.GetCurrencyRateFromContext(ctx, quote.Currency)

			cost := aggregatedLot.Cost * currencyRateByUse.PositionCost
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
				CurrencyConverted:  currencyRateByUse.ToCurrencyCode,
			}

			acc.positions = append(acc.positions, position)
			acc.positionSummaryBase = getPositionSummaryBase(position, acc.positionSummaryBase, currencyRateByUse)
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

func getPositionSummaryBase(position Position, acc positionSummaryBase, currencyRateByUse currency.CurrencyRateByUse) positionSummaryBase {
	acc.value += (position.Value * currencyRateByUse.SummaryValue)
	acc.cost += (position.Cost * currencyRateByUse.SummaryCost)
	acc.dayChange += (position.DayChange)
	return acc
}

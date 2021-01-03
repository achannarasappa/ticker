package position

import (
	. "ticker-tape/internal/quote"

	"github.com/novalagung/gubrak/v2"
)

type Position struct {
	AggregatedLot
	Value            float64
	DayChange        float64
	DayChangePercent float64
}

func GetPositions(quotes []Quote) map[string]Position {

	aggregatedLots := aggregateLotsBySymbol(lots)

	positions := gubrak.
		From(quotes).
		Reduce(func(acc []Position, quote Quote) []Position {
			if _, ok := aggregatedLots[quote.Symbol]; ok {
				return append(acc, Position{
					AggregatedLot:    aggregatedLots[quote.Symbol],
					Value:            quote.Price * aggregatedLots[quote.Symbol].Quantity,
					DayChange:        quote.Change * aggregatedLots[quote.Symbol].Quantity,
					DayChangePercent: quote.ChangePercent * aggregatedLots[quote.Symbol].Quantity,
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

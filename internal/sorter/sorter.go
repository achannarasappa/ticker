package sorter

import (
	"ticker/internal/position"
	"ticker/internal/quote"

	"github.com/novalagung/gubrak/v2"
)

type Sorter func(quotes []quote.Quote, positions map[string]position.Position) []quote.Quote

func NewSorter(sort string) Sorter {
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	} else {
		return defaultSorter
	}
}

var sortDict = map[string]Sorter{
	"alpha": func(quotes []quote.Quote, positions map[string]position.Position) []quote.Quote {
		if len(quotes) <= 0 {
			return quotes
		}

		result := gubrak.
			From(quotes).
			OrderBy(func(v quote.Quote) string {
				return v.Symbol
			}).
			Result()

		return (result).([]quote.Quote)
	},
	"position": func(quotes []quote.Quote, positions map[string]position.Position) []quote.Quote {
		if len(quotes) <= 0 {
			return quotes
		}

		activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

		cActiveQuotes := gubrak.From(activeQuotes)
		cInactiveQuotes := gubrak.From(inactiveQuotes)

		positionsSorter := func(v quote.Quote) float64 {
			return positions[v.Symbol].Value
		}

		cActiveQuotes.OrderBy(positionsSorter, false)
		cInactiveQuotes.OrderBy(positionsSorter, false)

		result := cActiveQuotes.
			Concat(cInactiveQuotes.Result()).
			Result()

		return (result).([]quote.Quote)
	},
}

func defaultSorter(quotes []quote.Quote, positions map[string]position.Position) []quote.Quote {
	if len(quotes) <= 0 {
		return quotes
	}

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	cActiveQuotes := gubrak.
		From(activeQuotes)

	cActiveQuotes.OrderBy(func(v quote.Quote) float64 {
		return v.ChangePercent
	}, false)

	result := cActiveQuotes.
		Concat(inactiveQuotes).
		Result()

	return (result).([]quote.Quote)
}

func splitActiveQuotes(quotes []quote.Quote) (interface{}, interface{}) {
	activeQuotes, inactiveQuotes, _ := gubrak.
		From(quotes).
		Partition(func(v quote.Quote) bool {
			return v.IsActive
		}).
		ResultAndError()

	return activeQuotes, inactiveQuotes
}

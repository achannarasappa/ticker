package sorter

import (
	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"

	"github.com/novalagung/gubrak/v2"
)

type Sorter func(quotes []Quote, positions map[string]Position) []Quote

func NewSorter(sort string) Sorter {
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	} else {
		return defaultSorter
	}
}

var sortDict = map[string]Sorter{
	"alpha": func(quotes []Quote, positions map[string]Position) []Quote {
		if len(quotes) <= 0 {
			return quotes
		}

		result := gubrak.
			From(quotes).
			OrderBy(func(v Quote) string {
				return v.Symbol
			}).
			Result()

		return (result).([]Quote)
	},
	"position": func(quotes []Quote, positions map[string]Position) []Quote {
		if len(quotes) <= 0 {
			return quotes
		}

		activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

		cActiveQuotes := gubrak.From(activeQuotes)
		cInactiveQuotes := gubrak.From(inactiveQuotes)

		positionsSorter := func(v Quote) float64 {
			return positions[v.Symbol].Value
		}

		cActiveQuotes.OrderBy(positionsSorter, false)
		cInactiveQuotes.OrderBy(positionsSorter, false)

		result := cActiveQuotes.
			Concat(cInactiveQuotes.Result()).
			Result()

		return (result).([]Quote)
	},
}

func defaultSorter(quotes []Quote, positions map[string]Position) []Quote {
	if len(quotes) <= 0 {
		return quotes
	}

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	cActiveQuotes := gubrak.
		From(activeQuotes)

	cActiveQuotes.OrderBy(func(v Quote) float64 {
		return v.ChangePercent
	}, false)

	result := cActiveQuotes.
		Concat(inactiveQuotes).
		Result()

	return (result).([]Quote)
}

func splitActiveQuotes(quotes []Quote) (interface{}, interface{}) {
	activeQuotes, inactiveQuotes, _ := gubrak.
		From(quotes).
		Partition(func(v Quote) bool {
			return v.IsActive
		}).
		ResultAndError()

	return activeQuotes, inactiveQuotes
}

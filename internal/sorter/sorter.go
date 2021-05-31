package sorter

import (
	"sort"

	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"
)

type Sorter func(quotes []Quote, positions map[string]Position) []Quote

func NewSorter(sort string) Sorter {
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	} else {
		return sortByChange
	}
}

var sortDict = map[string]Sorter{
	"alpha": sortByAlpha,
	"value": sortByValue,
	"user":  sortByUser,
}

func sortByUser(q []Quote, positions map[string]Position) []Quote {

	quoteCount := len(q)

	if quoteCount <= 0 {
		return q
	}

	quotes := make([]Quote, quoteCount)
	copy(quotes, q)

	sort.SliceStable(quotes, func(i, j int) bool {

		prevIndex := quoteCount
		nextIndex := quoteCount

		if position, ok := positions[quotes[i].Symbol]; ok {
			prevIndex = position.AggregatedLot.OrderIndex
		}

		if position, ok := positions[quotes[j].Symbol]; ok {
			nextIndex = position.AggregatedLot.OrderIndex
		}

		return nextIndex > prevIndex
	})

	return quotes

}

func sortByAlpha(q []Quote, positions map[string]Position) []Quote {

	quoteCount := len(q)

	if quoteCount <= 0 {
		return q
	}

	quotes := make([]Quote, quoteCount)
	copy(quotes, q)

	sort.SliceStable(quotes, func(i, j int) bool {
		return quotes[j].Symbol > quotes[i].Symbol
	})

	return quotes
}

func sortByValue(q []Quote, positions map[string]Position) []Quote {

	quoteCount := len(q)

	if quoteCount <= 0 {
		return q
	}

	quotes := make([]Quote, quoteCount)
	copy(quotes, q)

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	sort.SliceStable(inactiveQuotes, func(i, j int) bool {
		return positions[inactiveQuotes[j].Symbol].Value < positions[inactiveQuotes[i].Symbol].Value
	})

	sort.SliceStable(activeQuotes, func(i, j int) bool {
		return positions[activeQuotes[j].Symbol].Value < positions[activeQuotes[i].Symbol].Value
	})

	return append(activeQuotes, inactiveQuotes...)
}

func sortByChange(q []Quote, positions map[string]Position) []Quote {

	quoteCount := len(q)

	if quoteCount <= 0 {
		return q
	}

	quotes := make([]Quote, quoteCount)
	copy(quotes, q)

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	sort.SliceStable(activeQuotes, func(i, j int) bool {
		return activeQuotes[j].ChangePercent < activeQuotes[i].ChangePercent
	})

	return append(activeQuotes, inactiveQuotes...)

}

func splitActiveQuotes(quotes []Quote) ([]Quote, []Quote) {

	activeQuotes := make([]Quote, 0)
	inactiveQuotes := make([]Quote, 0)

	for _, quote := range quotes {
		if quote.IsActive {
			activeQuotes = append(activeQuotes, quote)
		} else {
			inactiveQuotes = append(inactiveQuotes, quote)
		}
	}

	return activeQuotes, inactiveQuotes
}

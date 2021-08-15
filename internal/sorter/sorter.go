package sorter

import (
	"sort"

	p "github.com/achannarasappa/ticker/internal/position"
	q "github.com/achannarasappa/ticker/internal/quote"
)

// Sorter represents a function that sorts quotes
type Sorter func(quotes []q.Quote, positions map[string]p.Position) []q.Quote

// NewSorter creates a sorting function
func NewSorter(sort string) Sorter {
	if sorter, ok := sortDict[sort]; ok {
		return sorter
	}
	return sortByChange
}

var sortDict = map[string]Sorter{
	"alpha": sortByAlpha,
	"value": sortByValue,
	"user":  sortByUser,
}

func sortByUser(quoteIn []q.Quote, positions map[string]p.Position) []q.Quote {

	quoteCount := len(quoteIn)

	if quoteCount <= 0 {
		return quoteIn
	}

	quotes := make([]q.Quote, quoteCount)
	copy(quotes, quoteIn)

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

func sortByAlpha(quoteIn []q.Quote, positions map[string]p.Position) []q.Quote {

	quoteCount := len(quoteIn)

	if quoteCount <= 0 {
		return quoteIn
	}

	quotes := make([]q.Quote, quoteCount)
	copy(quotes, quoteIn)

	sort.SliceStable(quotes, func(i, j int) bool {
		return quotes[j].Symbol > quotes[i].Symbol
	})

	return quotes
}

func sortByValue(quoteIn []q.Quote, positions map[string]p.Position) []q.Quote {

	quoteCount := len(quoteIn)

	if quoteCount <= 0 {
		return quoteIn
	}

	quotes := make([]q.Quote, quoteCount)
	copy(quotes, quoteIn)

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	sort.SliceStable(inactiveQuotes, func(i, j int) bool {
		return positions[inactiveQuotes[j].Symbol].Value < positions[inactiveQuotes[i].Symbol].Value
	})

	sort.SliceStable(activeQuotes, func(i, j int) bool {
		return positions[activeQuotes[j].Symbol].Value < positions[activeQuotes[i].Symbol].Value
	})

	return append(activeQuotes, inactiveQuotes...)
}

func sortByChange(quoteIn []q.Quote, positions map[string]p.Position) []q.Quote {

	quoteCount := len(quoteIn)

	if quoteCount <= 0 {
		return quoteIn
	}

	quotes := make([]q.Quote, quoteCount)
	copy(quotes, quoteIn)

	activeQuotes, inactiveQuotes := splitActiveQuotes(quotes)

	sort.SliceStable(activeQuotes, func(i, j int) bool {
		return activeQuotes[j].ChangePercent < activeQuotes[i].ChangePercent
	})

	sort.SliceStable(inactiveQuotes, func(i, j int) bool {
		return inactiveQuotes[j].ChangePercent < inactiveQuotes[i].ChangePercent
	})

	return append(activeQuotes, inactiveQuotes...)

}

func splitActiveQuotes(quotes []q.Quote) ([]q.Quote, []q.Quote) {

	activeQuotes := make([]q.Quote, 0)
	inactiveQuotes := make([]q.Quote, 0)

	for _, quote := range quotes {
		if quote.IsActive {
			activeQuotes = append(activeQuotes, quote)
		} else {
			inactiveQuotes = append(inactiveQuotes, quote)
		}
	}

	return activeQuotes, inactiveQuotes
}

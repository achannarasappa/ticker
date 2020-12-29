package watchlist

import (
	"math"
	"ticker-tape/internal/quote"
	. "ticker-tape/internal/ui/util"

	. "ticker-tape/internal/ui/util/text"

	"github.com/lucasb-eyer/go-colorful"
	. "github.com/novalagung/gubrak"
)

var (
	styleNeutral       = NewStyle("#d4d4d4", "", false)
	styleNeutralBold   = NewStyle("#d4d4d4", "", true)
	styleNeutralFaded  = NewStyle("#616161", "", false)
	styleNeutralBgBold = NewStyle("#4e4e4e", "#262626", true)
	stylePricePositive = newStyleFromGradient("#C6FF40", "#779929")
	stylePriceNegative = newStyleFromGradient("#FF7940", "#994926")
)

const (
	maxPercentChangeColorGradient = 10
)

type Model struct {
	Width  int
	Quotes []quote.Quote
}

// NewModel returns a model with default values.
func NewModel() Model {
	return Model{
		Width: 100,
	}
}

func (m Model) View() string {
	return watchlist(m.Quotes, m.Width)
}

func watchlist(q []quote.Quote, elementWidth int) string {
	quotes := sortQuotes(q)
	quoteSummaries := ""
	for _, quote := range quotes {
		quoteSummaries = quoteSummaries + "\n" + quoteSummary(quote, elementWidth)
	}
	return quoteSummaries
}

func quoteSummary(q quote.Quote, width int) string {

	return JoinText(
		Text(
			width,
			styleNeutralBold(q.Symbol),
			Right(
				Text(
					30,
					marketStateText(q),
					Right(
						styleNeutral(ConvertFloatToString(q.Price)),
					),
				),
			),
		),
		Text(
			width,
			styleNeutralFaded(q.ShortName),
			Right(
				priceText(q.Change, q.ChangePercent),
			),
		),
	)
}

func marketStateText(q quote.Quote) string {
	if q.IsRegularTradingSession || !q.IsActive {
		return ""
	}
	return styleNeutralBgBold(" " + q.MarketState + " ")
}

func priceText(change float64, changePercent float64) string {
	if change == 0.0 {
		return styleNeutralFaded("  " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	if change > 0.0 {
		return stylePricePositive(changePercent)("↑ " + ConvertFloatToString(change) + "  (" + ConvertFloatToString(changePercent) + "%)")
	}

	return stylePriceNegative(changePercent)("↓ " + ConvertFloatToString(change) + " (" + ConvertFloatToString(changePercent) + "%)")
}

func newStyleFromGradient(startColorHex string, endColorHex string) func(float64) func(string) string {
	c1, _ := colorful.Hex(startColorHex)
	c2, _ := colorful.Hex(endColorHex)

	return func(percent float64) func(string) string {
		normalizedPercent := getNormalizedPercentWithMax(percent, maxPercentChangeColorGradient)
		return NewStyle(c1.BlendHsv(c2, normalizedPercent).Hex(), "", false)
	}
}

// Normalize 0-100 percent with a maximum percent value
func getNormalizedPercentWithMax(percent float64, maxPercent float64) float64 {

	absolutePercent := math.Abs(percent)
	if absolutePercent >= maxPercent {
		return 1.0
	}
	return math.Abs(percent / maxPercent)

}

// Sort by change percent and keep all inactive quotes at the end
func sortQuotes(q []quote.Quote) []quote.Quote {
	if len(q) <= 0 {
		return q
	}

	activeQuotes, inactiveQuotes, _ := Partition(q, func(v quote.Quote) bool {
		return v.IsActive
	})

	sortedActiveQuotes, _ := OrderBy(activeQuotes, func(v quote.Quote) float64 {
		return v.ChangePercent
	}, false)

	concatQuotes, _ := Concat(sortedActiveQuotes, inactiveQuotes)

	return (concatQuotes).([]quote.Quote)
}

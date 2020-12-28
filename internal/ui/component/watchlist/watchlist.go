package watchlist

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui/util"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/muesli/reflow/ansi"
	. "github.com/novalagung/gubrak"
)

var (
	styleNeutral       = util.NewStyle("#d4d4d4", "", false)
	styleNeutralBold   = util.NewStyle("#d4d4d4", "", true)
	styleNeutralFaded  = util.NewStyle("#616161", "", false)
	stylePricePositive = newStyleFromGradient("#D8FF80", "#75BF00")
	stylePriceNegative = newStyleFromGradient("#FFA780", "#BF3900")
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

func quoteSummary(q quote.Quote, elementWidth int) string {

	firstLine := lineWithGap(
		styleNeutralBold(q.Symbol),
		styleNeutral(convertFloatToString(q.Price)),
		elementWidth,
	)
	secondLine := lineWithGap(
		styleNeutralFaded(q.ShortName),
		priceText(q.Change, q.ChangePercent),
		elementWidth,
	)

	return fmt.Sprintf("%s\n%s", firstLine, secondLine)
}

func priceText(change float64, changePercent float64) string {
	if change == 0.0 {
		return styleNeutralFaded("  " + convertFloatToString(change) + "  (" + convertFloatToString(changePercent) + "%)")
	}

	if change > 0.0 {
		return stylePricePositive(changePercent)("↑ " + convertFloatToString(change) + "  (" + convertFloatToString(changePercent) + "%)")
	}

	return stylePriceNegative(changePercent)("↓ " + convertFloatToString(change) + " (" + convertFloatToString(changePercent) + "%)")
}

// util
func lineWithGap(leftText string, rightText string, elementWidth int) string {
	innerGapWidth := elementWidth - ansi.PrintableRuneWidth(leftText) - ansi.PrintableRuneWidth(rightText)
	if innerGapWidth > 0 {
		return leftText + strings.Repeat(" ", innerGapWidth) + rightText
	}

	return leftText + " " + rightText
}

func convertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

func newStyleFromGradient(startColorHex string, endColorHex string) func(float64) func(string) string {
	c1, _ := colorful.Hex(startColorHex)
	c2, _ := colorful.Hex(endColorHex)

	return func(percent float64) func(string) string {
		normalizedPercent := getNormalizedPercentWithMax(percent, maxPercentChangeColorGradient)
		return util.NewStyle(c1.BlendHsv(c2, normalizedPercent).Hex(), "", false)
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

	sortedActiveQuotes, _ := SortBy(activeQuotes, func(v quote.Quote) float64 {
		return v.ChangePercent
	})

	concatQuotes, _ := Concat(sortedActiveQuotes, inactiveQuotes)

	return (concatQuotes).([]quote.Quote)
}

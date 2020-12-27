package watchlist

import (
	"fmt"
	"strconv"
	"strings"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui/util"

	"github.com/muesli/reflow/ansi"
)

var (
	styleNeutral       = util.NewStyle("#d4d4d4", "", false)
	styleNeutralFaded  = util.NewStyle("#7e8087", "", false)
	stylePricePositive = util.NewStyle("#d1ff82", "", false)
	stylePriceNegative = util.NewStyle("#ff8c82", "", false)
)

const (
	footerHeight = 1
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
	return quoteSummaries(m.Quotes, m.Width)
}

func quoteSummaries(q []quote.Quote, elementWidth int) string {
	quoteSummaries := ""
	for _, quote := range q {
		quoteSummaries = quoteSummaries + "\n" + quoteSummary(quote, elementWidth)
	}
	return quoteSummaries
}

func quoteSummary(q quote.Quote, elementWidth int) string {

	firstLine := lineWithGap(
		styleNeutral(q.Symbol),
		styleNeutral(convertFloatToString(q.RegularMarketPrice)),
		elementWidth,
	)
	secondLine := lineWithGap(
		styleNeutralFaded(q.ShortName),
		stylePricePositive("↑ "+convertFloatToString(q.RegularMarketChange)+" ("+convertFloatToString(q.RegularMarketChangePercent)+"%)"),
		elementWidth,
	)

	return fmt.Sprintf("%s\n%s", firstLine, secondLine)
}

// util
func lineWithGap(leftText string, rightText string, elementWidth int) string {
	innerGapWidth := elementWidth - ansi.PrintableRuneWidth(leftText) - ansi.PrintableRuneWidth(rightText)
	return leftText + strings.Repeat(" ", innerGapWidth) + rightText
}

func convertFloatToString(f float64) string {
	return strconv.FormatFloat(f, 'f', 2, 64)
}

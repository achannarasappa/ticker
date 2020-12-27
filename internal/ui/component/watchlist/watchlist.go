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
)

var (
	styleNeutral       = util.NewStyle("#d4d4d4", "", false)
	styleNeutralBold   = util.NewStyle("#d4d4d4", "", true)
	styleNeutralFaded  = util.NewStyle("#616161", "", false)
	stylePricePositive = newStyleFromGradient("#cae891", "#adff00")
	stylePriceNegative = newStyleFromGradient("#f5a07d", "#ff5200")
)

const (
	maxPercentChangeColorGradient = 10
)

const (
	PositivePriceChange = iota
	NegativePriceChange
	NeutralPriceChange
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
	quoteSummaries := ""
	for _, quote := range q {
		quoteSummaries = quoteSummaries + "\n" + quoteSummary(quote, elementWidth)
	}
	return quoteSummaries
}

func quoteSummary(q quote.Quote, elementWidth int) string {

	p := getPrice(q)

	firstLine := lineWithGap(
		styleNeutralBold(q.Symbol),
		styleNeutral(convertFloatToString(p.Price)),
		elementWidth,
	)
	secondLine := lineWithGap(
		styleNeutralFaded(q.ShortName),
		priceText(p.Change, p.ChangePercent),
		elementWidth,
	)

	return fmt.Sprintf("%s\n%s", firstLine, secondLine)
}

type quoteMeta struct {
	Price         float64
	Change        float64
	ChangePercent float64
}

func getPrice(q quote.Quote) quoteMeta {

	if q.MarketState == "REGULAR" {
		return quoteMeta{
			Price:         q.RegularMarketPrice,
			Change:        q.RegularMarketChange,
			ChangePercent: q.RegularMarketChangePercent,
		}
	}

	if q.MarketState == "POST" {
		return quoteMeta{
			Price:         q.PostMarketPrice,
			Change:        q.PostMarketChange,
			ChangePercent: q.PostMarketChangePercent,
		}
	}

	return quoteMeta{
		Price:         q.RegularMarketPrice,
		Change:        0.0,
		ChangePercent: 0.0,
	}

}

func priceText(change float64, changePercent float64) string {
	if change == 0.0 {
		return styleNeutralFaded("  " + convertFloatToString(change) + "  (" + convertFloatToString(changePercent) + "%)")
	}

	if change > 0.0 {
		return stylePricePositive(change)("↑ " + convertFloatToString(change) + "  (" + convertFloatToString(changePercent) + "%)")
	}

	return stylePriceNegative(change)("↓ " + convertFloatToString(change) + " (" + convertFloatToString(changePercent) + "%)")
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
	if absolutePercent > maxPercent {
		return 1
	}

	return percent / maxPercent

}

// Accepts a start color, end color, and percent and returns the color at the position between them
func getColorHex(startColorHex string, endColorHex string, normalizedPercent float64) string {

	c1, _ := colorful.Hex(startColorHex)
	c2, _ := colorful.Hex(endColorHex)

	return c1.BlendHsv(c2, normalizedPercent).Hex()

}

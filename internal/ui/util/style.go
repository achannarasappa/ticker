package util

import (
	"math"

	"github.com/lucasb-eyer/go-colorful"
	te "github.com/muesli/termenv"
)

const (
	maxPercentChangeColorGradient = 10
)

var (
	StyleNeutral       = NewStyle("#d0d0d0", "", false)
	StyleNeutralBold   = NewStyle("#d0d0d0", "", true)
	StyleNeutralLight  = NewStyle("#8a8a8a", "", false)
	StyleNeutralFaded  = NewStyle("#626262", "", false)
	StyleLine          = NewStyle("#3a3a3a", "", false)
	StyleTag           = NewStyle("#8a8a8a", "#303030", false)
	StyleTagEnd        = NewStyle("#303030", "#303030", false)
	StylePricePositive = NewStyleFromGradient("#C6FF40", "#779929")
	StylePriceNegative = NewStyleFromGradient("#FF7940", "#994926")
)

func NewStyle(fg string, bg string, bold bool) func(string) string {
	s := te.Style{}.Foreground(te.ColorProfile().Color(fg)).Background(te.ColorProfile().Color(bg))
	if bold {
		s = s.Bold()
	}
	return s.Styled
}

func NewStyleFromGradient(startColorHex string, endColorHex string) func(float64) func(string) string {
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

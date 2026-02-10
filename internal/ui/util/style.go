package util

import (
	"math"
	"regexp"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/lucasb-eyer/go-colorful"
	te "github.com/muesli/termenv"
)

const (
	maxPercentChangeColorGradient = 10
	strongChangeThreshold         = 10.0
	mediumChangeThreshold         = 5.0
)

//nolint:gochecknoglobals
var (
	p = te.ColorProfile()
)

var defaultColorScheme = c.ConfigColorScheme{
	Text:          "#d0d0d0",
	TextLight:     "#8a8a8a",
	TextLabel:     "#626262",
	TextLine:      "#3a3a3a",
	BackgroundTag: "#303030",
	TextTag:       "#8a8a8a",
	PriceColorScheme: c.ConfigPriceColorSchemeGroup{
		Light: c.ConfigPriceColorScheme{
			PositiveStart: "#22C55E",
			PositiveEnd:   "#15803D",
			NegativeStart: "#EF4444",
			NegativeEnd:   "#B91C1C",
		},
		Dark: c.ConfigPriceColorScheme{
			PositiveStart: "#C6FF40",
			PositiveEnd:   "#779929",
			NegativeStart: "#FF7940",
			NegativeEnd:   "#994926",
		},
	},
}

// NewStyle creates a new predefined style function
func NewStyle(fg string, bg string, bold bool) func(string) string {
	s := te.Style{}.Foreground(p.Color(fg)).Background(p.Color(bg))
	if bold {
		s = s.Bold()
	}

	return s.Styled
}

func getStylePriceFn(colorSchemeGroup c.ConfigPriceColorSchemeGroup) func(float64, string) string {
	return func(percent float64, text string) string { //nolint:cyclop

		colorScheme := colorSchemeGroup.Light
		if te.HasDarkBackground() {
			colorScheme = colorSchemeGroup.Dark
		}

		out := te.String(text)

		if percent == 0.0 {
			return out.Foreground(p.Color("241")).String()
		}

		if p == te.TrueColor && percent > 0.0 {
			return newStyleFromGradient(colorScheme.PositiveStart, colorScheme.PositiveEnd)(percent, text)
		}

		if p == te.TrueColor && percent < 0.0 {
			return newStyleFromGradient(colorScheme.NegativeStart, colorScheme.NegativeEnd)(percent, text)
		}

		if percent > strongChangeThreshold {
			return out.Foreground(p.Color("70")).String()
		}

		if percent > mediumChangeThreshold {
			return out.Foreground(p.Color("76")).String()
		}

		if percent > 0.0 {
			return out.Foreground(p.Color("82")).String()
		}

		if percent < strongChangeThreshold*-1 {
			return out.Foreground(p.Color("124")).String()
		}

		if percent < mediumChangeThreshold*-1 {
			return out.Foreground(p.Color("160")).String()
		}

		return out.Foreground(p.Color("196")).String()
	}
}

func newStyleFromGradient(startColorHex string, endColorHex string) func(float64, string) string {
	c1, _ := colorful.Hex(startColorHex)
	c2, _ := colorful.Hex(endColorHex)

	return func(percent float64, text string) string {
		normalizedPercent := getNormalizedPercentWithMax(percent, maxPercentChangeColorGradient)
		textColor := p.Color(c1.BlendHsv(c2, normalizedPercent).Hex())

		return te.String(text).Foreground(textColor).String()
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

// GetColorScheme generates a color scheme based on user defined colors or defaults
func GetColorScheme(colorScheme c.ConfigColorScheme) c.Styles {
	priceColorScheme := getPriceColorSchemeGroup(colorScheme)
	return c.Styles{
		Text: NewStyle(
			getColorOrDefault(colorScheme.Text, defaultColorScheme.Text),
			"",
			false,
		),
		TextLight: NewStyle(
			getColorOrDefault(colorScheme.TextLight, defaultColorScheme.TextLight),
			"",
			false,
		),
		TextBold: NewStyle(
			getColorOrDefault(colorScheme.Text, defaultColorScheme.Text),
			"",
			true,
		),
		TextLabel: NewStyle(
			getColorOrDefault(colorScheme.TextLabel, defaultColorScheme.TextLabel),
			"",
			false,
		),
		TextLine: NewStyle(
			getColorOrDefault(colorScheme.TextLine, defaultColorScheme.TextLine),
			"",
			false,
		),
		TextPrice: getStylePriceFn(priceColorScheme),
		Tag: NewStyle(
			getColorOrDefault(colorScheme.TextTag, defaultColorScheme.TextLight),
			getColorOrDefault(colorScheme.BackgroundTag, defaultColorScheme.BackgroundTag),
			false,
		),
	}
}

// getPriceColorSchemeGroup returns the color scheme group for price changes based
// on user defined colors or defaults for light and dark mode
func getPriceColorSchemeGroup(colorScheme c.ConfigColorScheme) c.ConfigPriceColorSchemeGroup {
	lightSource := &colorScheme.PriceColorScheme.Light
	lightDefaults := &defaultColorScheme.PriceColorScheme.Light
	darkSource := &colorScheme.PriceColorScheme.Dark
	darkDefaults := &defaultColorScheme.PriceColorScheme.Dark

	return c.ConfigPriceColorSchemeGroup{
		Light: c.ConfigPriceColorScheme{
			PositiveStart: getColorOrDefault(lightSource.PositiveStart, lightDefaults.PositiveStart),
			PositiveEnd:   getColorOrDefault(lightSource.PositiveEnd, lightDefaults.PositiveEnd),
			NegativeStart: getColorOrDefault(lightSource.NegativeStart, lightDefaults.NegativeStart),
			NegativeEnd:   getColorOrDefault(lightSource.NegativeEnd, lightDefaults.NegativeEnd),
		},
		Dark: c.ConfigPriceColorScheme{
			PositiveStart: getColorOrDefault(darkSource.PositiveStart, darkDefaults.PositiveStart),
			PositiveEnd:   getColorOrDefault(darkSource.PositiveEnd, darkDefaults.PositiveEnd),
			NegativeStart: getColorOrDefault(darkSource.NegativeStart, darkDefaults.NegativeStart),
			NegativeEnd:   getColorOrDefault(darkSource.NegativeEnd, darkDefaults.NegativeEnd),
		},
	}
}

func getColorOrDefault(colorConfig string, colorDefault string) string {
	re := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)

	if len(re.FindStringIndex(colorConfig)) > 0 {
		return colorConfig
	}

	return colorDefault
}

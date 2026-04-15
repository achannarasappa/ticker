package util

import (
	"math"
	"regexp"
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/lucasb-eyer/go-colorful"
	te "github.com/muesli/termenv"
)

const (
	maxPercentChangeColorGradient = 10
)

//nolint:gochecknoglobals
var (
	p                  = te.ColorProfile()
	stylePricePositive = newStyleFromGradient("#C6FF40", "#779929")
	stylePriceNegative = newStyleFromGradient("#FF7940", "#994926")
)

// NewStyle creates a new predefined style function
func NewStyle(fg string, bg string, bold bool) func(string) string {
	s := te.Style{}.Foreground(p.Color(fg)).Background(p.Color(bg))
	if bold {
		s = s.Bold()
	}

	return s.Styled
}

func stylePrice(percent float64, text string) string { //nolint:cyclop

	out := te.String(text)

	if percent == 0.0 {
		return out.Foreground(p.Color("241")).String()
	}

	if p == te.TrueColor && percent > 0.0 {
		return stylePricePositive(percent, text)
	}

	if p == te.TrueColor && percent < 0.0 {
		return stylePriceNegative(percent, text)
	}

	if percent > 10.0 {
		return out.Foreground(p.Color("70")).String()
	}

	if percent > 5 {
		return out.Foreground(p.Color("76")).String()
	}

	if percent > 0.0 {
		return out.Foreground(p.Color("82")).String()
	}

	if percent < -10.0 {
		return out.Foreground(p.Color("124")).String()
	}

	if percent < -5.0 {
		return out.Foreground(p.Color("160")).String()
	}

	return out.Foreground(p.Color("196")).String()

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

	return c.Styles{
		Text: NewStyle(
			getColorOrDefault(colorScheme.Text, "#d0d0d0"),
			"",
			false,
		),
		TextLight: NewStyle(
			getColorOrDefault(colorScheme.TextLight, "#8a8a8a"),
			"",
			false,
		),
		TextBold: NewStyle(
			getColorOrDefault(colorScheme.Text, "#d0d0d0"),
			"",
			true,
		),
		TextLabel: NewStyle(
			getColorOrDefault(colorScheme.TextLabel, "#626262"),
			"",
			false,
		),
		TextLine: NewStyle(
			getColorOrDefault(colorScheme.TextLine, "#3a3a3a"),
			"",
			false,
		),
		TextPrice: stylePrice,
		Tag: NewStyle(
			getColorOrDefault(colorScheme.TextTag, "#8a8a8a"),
			getColorOrDefault(colorScheme.BackgroundTag, "#303030"),
			false,
		),
	}

}

func getColorOrDefault(colorConfig string, colorDefault string) string {
	re := regexp.MustCompile(`^#(?:[0-9a-fA-F]{3}){1,2}$`)

	if len(re.FindStringIndex(colorConfig)) > 0 {
		return colorConfig
	}

	return colorDefault
}

// GetRowAlternateBackground returns the resolved background color hex for alternate rows
func GetRowAlternateBackground(colorScheme c.ConfigColorScheme) string {
	return getColorOrDefault(colorScheme.BackgroundRow, "#303030")
}

// GetRowAlternateBackgroundColor returns the alternate row background color when the feature
// is enabled in the config, or an empty string when the feature is disabled.
func GetRowAlternateBackgroundColor(config c.Config) string {
	if !config.ShowAlternateRowBackground {
		return ""
	}

	return GetRowAlternateBackground(config.ColorScheme)
}

// ApplyBackground applies a background color to a rendered terminal string,
// including re-applying after any terminal reset sequences so that fill spaces
// between styled chunks also receive the background color.
func ApplyBackground(text string, bgHex string) string {
	if bgHex == "" {
		return text
	}

	bgColor := p.Color(bgHex)

	// te.Style{}.Background(color).Styled("") returns "\x1b[48;...m\x1b[0m" or "" for ASCII profile
	openCode := te.Style{}.Background(bgColor).Styled("")
	openBgCode := strings.TrimSuffix(openCode, "\x1b[0m")

	if openBgCode == "" {
		// ASCII profile or unknown color - no background available
		return text
	}

	const resetCode = "\x1b[0m"

	// Prepend background, and re-apply it after every reset in the text
	result := openBgCode + strings.ReplaceAll(text, resetCode, resetCode+openBgCode)

	return result + resetCode
}

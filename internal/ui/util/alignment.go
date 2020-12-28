package util

import (
	"strings"

	"github.com/muesli/reflow/ansi"
)

func LineWithGap(leftText string, rightText string, elementWidth int) string {
	innerGapWidth := elementWidth - ansi.PrintableRuneWidth(leftText) - ansi.PrintableRuneWidth(rightText)
	if innerGapWidth > 0 {
		return leftText + strings.Repeat(" ", innerGapWidth) + rightText
	}

	return leftText + " " + rightText
}

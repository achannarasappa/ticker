package text

import (
	"math"
	"strings"

	"github.com/muesli/reflow/ansi"
)

type Element func(int) string
type ElementText func(string) Element
type ElementTextSegment func(int, int) string

func getElementWidth(widthTotal int, count int) (int, int) {
	remainder := widthTotal % count
	width := int(math.Floor(float64(widthTotal) / float64(count)))

	return width, remainder
}

func Text(width int, arguments ...interface{}) string {

	widthElementWithoutRemainder, remainder := getElementWidth(width, len(arguments))
	var elementString string
	for i, argument := range arguments {

		widthElement := widthElementWithoutRemainder
		if i < remainder {
			widthElement = widthElementWithoutRemainder + 1
		}

		switch element := argument.(type) {
		case Element:
			elementString = elementString + element(widthElement)
		case string:
			elementString = elementString + Left(element)(widthElement)
		}
	}

	return elementString

}

func Center(text string) Element {

	return func(width int) string {
		gapWidth := width - ansi.PrintableRuneWidth(text)
		if gapWidth > 0 {
			return strings.Repeat(" ", gapWidth) + text
		}

		return text
	}
}

func Right(text string) Element {

	return func(width int) string {
		gapWidth := width - ansi.PrintableRuneWidth(text)
		if gapWidth > 0 {
			return strings.Repeat(" ", gapWidth) + text
		}

		return text
	}
}

func Left(text string) Element {

	return func(width int) string {
		gapWidth := width - ansi.PrintableRuneWidth(text)
		if gapWidth > 0 {
			return text + strings.Repeat(" ", gapWidth)
		}

		return text
	}
}

func JoinText(texts ...string) string {
	return strings.Join(
		texts,
		"\n",
	)
}

// func getElementWidthWithEvenSpacing(width int, textSegmentCount int) int {
// 	return int(math.Floor(float64(width) / float64(textSegmentCount)))
// }

package text

import (
	"math"
	"strings"

	"github.com/muesli/reflow/ansi"
)

func getElementWidth(widthTotal int, count int) (int, int) {
	remainder := widthTotal % count
	width := int(math.Floor(float64(widthTotal) / float64(count)))

	return width, remainder
}

type TextAlign int

const (
	LeftAlign TextAlign = iota
	RightAlign
)

func (ta TextAlign) String() string {
	return [...]string{"LeftAlign", "RightAlign"}[ta]
}

type Cell struct {
	Text  string
	Width int
	Align TextAlign
}

func Line(width int, cells ...Cell) string {

	widthFlex := width
	var widthFlexCells []*int

	for i, cell := range cells {
		if cell.Width <= 0 {
			widthFlexCells = append(widthFlexCells, &cells[i].Width)
			continue
		}
		widthFlex -= cell.Width
	}

	widthWithoutRemainder, remainder := getElementWidth(widthFlex, len(widthFlexCells))
	for i := range widthFlexCells {

		*widthFlexCells[i] = widthWithoutRemainder
		if i < remainder {
			*widthFlexCells[i] = widthWithoutRemainder + 1
		}
	}

	var gridLine string
	for _, cell := range cells {

		textWidth := ansi.PrintableRuneWidth(cell.Text)
		if textWidth > cell.Width {
			cell.Text = cell.Text[:cell.Width]
			textWidth = cell.Width
		}

		if cell.Align == RightAlign {
			gridLine += strings.Repeat(" ", cell.Width-textWidth) + cell.Text
			continue
		}

		gridLine += cell.Text + strings.Repeat(" ", cell.Width-textWidth)
	}
	return gridLine

}

func JoinLines(texts ...string) string {
	return strings.Join(
		texts,
		"\n",
	)
}

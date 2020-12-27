package util

import (
	te "github.com/muesli/termenv"
)

func NewStyle(fg string, bg string, bold bool) func(string) string {
	s := te.Style{}.Foreground(te.ColorProfile().Color(fg)).Background(te.ColorProfile().Color(bg))
	if bold {
		s = s.Bold()
	}
	return s.Styled
}

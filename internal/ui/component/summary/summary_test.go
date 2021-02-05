package summary_test

import (
	"strings"

	"github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/ui/component/summary"

	"github.com/acarl005/stripansi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func removeFormatting(text string) string {
	return stripansi.Strip(text)
}

var _ = Describe("Summary", func() {

	It("should render a summary", func() {
		m := NewModel()
		m.Summary = position.PositionSummary{
			Value:            10000,
			Cost:             1000,
			Change:           9000,
			DayChange:        100.0,
			ChangePercent:    1000.0,
			DayChangePercent: 10.0,
		}
		Expect(removeFormatting(m.View())).To(Equal(strings.Join([]string{
			"Day: ↑ 100.00 (10.00%) • Change: ↑ 9000.00 (1000.00%) • Value: 10000.00",
			"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
		}, "\n")))
	})

	When("no quotes are set", func() {
		It("should render an empty summary", func() {
			m := NewModel()
			Expect(removeFormatting(m.View())).To(Equal(strings.Join([]string{
				"Day: 0.00 (0.00%) • Change: 0.00 (0.00%) • Value: ",
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
			}, "\n")))
		})
	})

	When("the window width is less than the minimum", func() {
		It("should render an empty summary", func() {
			m := NewModel()
			m.Width = 10
			Expect(m.View()).To(Equal(""))
		})
	})
})

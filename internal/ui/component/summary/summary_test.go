package summary_test

import (
	"strings"

	"github.com/achannarasappa/ticker/v5/internal/asset"
	c "github.com/achannarasappa/ticker/v5/internal/common"
	. "github.com/achannarasappa/ticker/v5/internal/ui/component/summary"

	"github.com/acarl005/stripansi"
	tea "github.com/charmbracelet/bubbletea"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func removeFormatting(text string) string {
	return stripansi.Strip(text)
}

var _ = Describe("Summary", func() {

	ctxFixture := c.Context{Reference: c.Reference{Styles: c.Styles{
		Text:      func(v string) string { return v },
		TextLight: func(v string) string { return v },
		TextLabel: func(v string) string { return v },
		TextBold:  func(v string) string { return v },
		TextLine:  func(v string) string { return v },
		TextPrice: func(percent float64, text string) string { return text },
		Tag:       func(v string) string { return v },
	}}}

	When("the change is positive", func() {
		It("should render a summary with up arrow", func() {
			m := NewModel(ctxFixture)
			m, _ = m.Update(tea.WindowSizeMsg{Width: 120})
			m, _ = m.Update(SetSummaryMsg(asset.HoldingSummary{
				Value: 10000,
				Cost:  1000,
				DayChange: c.HoldingChange{
					Amount:  100.0,
					Percent: 10.0,
				},
				TotalChange: c.HoldingChange{
					Amount:  9000,
					Percent: 1000.0,
				},
			}))
			Expect(removeFormatting(m.View())).To(Equal(strings.Join([]string{
				"Day Change: ↑ 100.00 (10.00%) • Change: ↑ 9,000.00 (1,000.00%)  • Value: 10,000.00  • Cost: 1,000.00  ",
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
			}, "\n")))
		})
	})

	When("the change is negative", func() {
		It("should render a summary with down arrow", func() {
			m := NewModel(ctxFixture)
			m, _ = m.Update(tea.WindowSizeMsg{Width: 120})
			m, _ = m.Update(SetSummaryMsg(asset.HoldingSummary{
				Value: 1000,
				Cost:  10000,
				DayChange: c.HoldingChange{
					Amount:  -100.0,
					Percent: -10.0,
				},
				TotalChange: c.HoldingChange{
					Amount:  -9000,
					Percent: -1000.0,
				},
			}))
			Expect(removeFormatting(m.View())).To(Equal(strings.Join([]string{
				"Day Change: ↓ -100.00 (-10.00%) • Change: ↓ -9,000.00 (-1,000.00%)  • Value: 1,000.00  • Cost: 10,000.00",
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
			}, "\n")))
		})
	})

	When("no quotes are set", func() {
		It("should render an empty summary", func() {
			m := NewModel(ctxFixture)
			Expect(removeFormatting(m.View())).To(Equal(strings.Join([]string{
				"Day Change: 0.00 (0.00%) • Change: 0.00 (0.00%)  • Value: 0.00  • Cost: 0.00 ",
				"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━",
			}, "\n")))
		})
	})

	When("the window width is less than the minimum", func() {
		It("should render an empty summary", func() {
			m := NewModel(ctxFixture)
			m, _ = m.Update(tea.WindowSizeMsg{Width: 10})
			Expect(m.View()).To(Equal(""))
		})
	})
})

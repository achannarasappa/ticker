package row_test

import (
	"strings"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	"github.com/achannarasappa/ticker/v5/internal/ui/component/watchlist/row"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"regexp"
)

var styles = c.Styles{
	Text:      func(v string) string { return v },
	TextLight: func(v string) string { return v },
	TextLabel: func(v string) string { return v },
	TextBold:  func(v string) string { return v },
	TextLine:  func(v string) string { return v },
	TextPrice: func(percent float64, text string) string { return text },
	Tag:       func(v string) string { return v },
}

var _ = Describe("Row", func() {

	Describe("Update", func() {

		Describe("UpdateAssetMsg", func() {

			When("the price has not changed or the symbol has changed", func() {

				It("should update the asset", func() {

					inputRow := row.New(row.Config{
						Styles: styles,
						Asset: &c.Asset{
							Symbol: "AAPL",
							QuotePrice: c.QuotePrice{
								Price: 150.00,
							},
						},
					})

					outputRow, _ := inputRow.Update(row.UpdateAssetMsg(&c.Asset{
						Symbol: "AAPL",
						QuotePrice: c.QuotePrice{
							Price: 150.00,
						},
					}))

					Expect(outputRow.View()).To(Equal(inputRow.View()))
				})

				When("the price has changed and symbol is the same", func() {

					When("the price has increased", func() {

						It("should animate the price increase", func() {
							stripANSI := func(str string) string {
								re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
								return re.ReplaceAllString(str, "")
							}

							inputRow := row.New(row.Config{
								ID:     1,
								Styles: styles,
								Asset: &c.Asset{
									Symbol: "AAPL",
									QuotePrice: c.QuotePrice{
										Price: 150.00,
									},
								},
							})

							// First update to trigger animation
							outputRow, cmd := inputRow.Update(row.UpdateAssetMsg(&c.Asset{
								Symbol: "AAPL",
								QuotePrice: c.QuotePrice{
									Price: 151.00,
								},
							}))

							view := stripANSI(outputRow.View())
							Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
							Expect(view).To(ContainSubstring("151.00"), "output was: %q", view)
							Expect(cmd).ToNot(BeNil())

							// Simulate frame updates
							for i := 0; i < 4; i++ {
								outputRow, cmd = outputRow.Update(row.FrameMsg(1))
								Expect(cmd).ToNot(BeNil())
							}

							// Final frame should have no animation
							outputRow, cmd = outputRow.Update(row.FrameMsg(1))

							view = stripANSI(outputRow.View())
							Expect(cmd).To(BeNil(), "expected cmd to be nil after final frame, got: %v", cmd)
							Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
							Expect(view).To(ContainSubstring("151.00"), "output was: %q", view)
						})

					})

					When("the price has decreased", func() {

						It("should animate the price decrease", func() {
							stripANSI := func(str string) string {
								re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
								return re.ReplaceAllString(str, "")
							}

							inputRow := row.New(row.Config{
								ID:     1,
								Styles: styles,
								Asset: &c.Asset{
									Symbol: "AAPL",
									QuotePrice: c.QuotePrice{
										Price: 151.00,
									},
								},
							})

							// First update to trigger animation
							outputRow, cmd := inputRow.Update(row.UpdateAssetMsg(&c.Asset{
								Symbol: "AAPL",
								QuotePrice: c.QuotePrice{
									Price: 150.00,
								},
							}))

							view := stripANSI(outputRow.View())
							Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
							Expect(view).To(ContainSubstring("150.00"), "output was: %q", view)
							Expect(cmd).ToNot(BeNil())

							// Simulate frame updates
							for i := 0; i < 4; i++ {
								outputRow, cmd = outputRow.Update(row.FrameMsg(1))
								Expect(cmd).ToNot(BeNil())
							}

							// Final frame should have no animation
							outputRow, cmd = outputRow.Update(row.FrameMsg(1))

							view = stripANSI(outputRow.View())
							Expect(cmd).To(BeNil(), "expected cmd to be nil after final frame, got: %v", cmd)
							Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
							Expect(view).To(ContainSubstring("150.00"), "output was: %q", view)
						})

					})

					When("the number of digits in the new and old price is different", func() {

						It("should animate the entire price", func() {
							stripANSI := func(str string) string {
								re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
								return re.ReplaceAllString(str, "")
							}

							inputRow := row.New(row.Config{
								ID:     1,
								Styles: styles,
								Asset: &c.Asset{
									Symbol: "AAPL",
									QuotePrice: c.QuotePrice{
										Price: 150.00,
									},
								},
							})

							// First update to trigger animation with different number of digits
							outputRow, cmd := inputRow.Update(row.UpdateAssetMsg(&c.Asset{
								Symbol: "AAPL",
								QuotePrice: c.QuotePrice{
									Price: 1500.00,
								},
							}))

							view := stripANSI(outputRow.View())
							Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
							Expect(view).To(ContainSubstring("1,500.00"), "output was: %q", view)
							Expect(cmd).ToNot(BeNil())

							// Simulate frame updates
							for i := 0; i < 4; i++ {
								outputRow, cmd = outputRow.Update(row.FrameMsg(1))
								Expect(cmd).ToNot(BeNil())
							}

							// Final frame should have no animation
							outputRow, cmd = outputRow.Update(row.FrameMsg(1))

							view = stripANSI(outputRow.View())
							Expect(cmd).To(BeNil(), "expected cmd to be nil after final frame, got: %v", cmd)
							Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
							Expect(view).To(ContainSubstring("1,500.00"), "output was: %q", view)
						})

					})

				})

			})

		})

		Describe("FrameMsg", func() {

			When("the message is for a different row", func() {

				It("should not animate the price", func() {
					stripANSI := func(str string) string {
						re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
						return re.ReplaceAllString(str, "")
					}

					inputRow := row.New(row.Config{
						ID:     1,
						Styles: styles,
						Asset: &c.Asset{
							Symbol: "AAPL",
							QuotePrice: c.QuotePrice{
								Price: 150.00,
							},
						},
					})

					// First update to trigger animation
					outputRow, cmd := inputRow.Update(row.UpdateAssetMsg(&c.Asset{
						Symbol: "AAPL",
						QuotePrice: c.QuotePrice{
							Price: 151.00,
						},
					}))

					// Send frame message for a different row ID
					outputRow, cmd = outputRow.Update(row.FrameMsg(2))

					view := stripANSI(outputRow.View())
					Expect(cmd).To(BeNil(), "expected cmd to be nil for different row ID")
					Expect(view).To(ContainSubstring("AAPL"), "output was: %q", view)
					Expect(view).To(ContainSubstring("151.00"), "output was: %q", view)
				})

			})

		})

		Describe("SetCellWidthsMsg", func() {

			It("should update the width and cell widths", func() {
				asset := &c.Asset{
					Symbol: "AAPL",
					QuotePrice: c.QuotePrice{
						Price: 150.00,
					},
				}

				inputRow := row.New(row.Config{
					Styles: styles,
					Asset:  asset,
				})

				expectedCellWidths := row.CellWidthsContainer{
					PositionLength:        10,
					QuoteLength:           8,
					WidthQuote:            12,
					WidthQuoteExtended:    15,
					WidthQuoteRange:       20,
					WidthPosition:         12,
					WidthPositionExtended: 15,
					WidthVolumeMarketCap:  15,
				}
				expectedWidth := 100

				outputRow, cmd := inputRow.Update(row.SetCellWidthsMsg{
					Width:      expectedWidth,
					CellWidths: expectedCellWidths,
				})

				Expect(cmd).To(BeNil())

				// Verify the width is applied by checking the rendered output
				view := outputRow.View()
				lines := strings.Split(view, "\n")
				Expect(lines[0]).To(HaveLen(expectedWidth))
			})

		})

	})

})

package watchlist_test

import (
	"fmt"
	"strings"

	"github.com/acarl005/stripansi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"

	. "ticker/internal/position"
	. "ticker/internal/quote"
	. "ticker/internal/ui/component/watchlist"
)

func removeFormatting(text string) string {
	return stripansi.Strip(text)
}

var _ = Describe("Watchlist", func() {
	describe := func(desc string) func(bool, bool, bool, float64, Position, string) string {
		return func(isActive bool, isRegularTradingSession bool, showTotals bool, change float64, position Position, expected string) string {
			return fmt.Sprintf("%s expected:%s", desc, expected)
		}
	}

	DescribeTable("should render a watchlist",
		func(isActive bool, isRegularTradingSession bool, showTotals bool, change float64, position Position, expected string) {

			var positionMap map[string]Position
			if (position == Position{}) {
				positionMap = map[string]Position{}
			} else {
				positionMap = map[string]Position{
					"AAPL": position,
				}
			}

			m := NewModel(false, false, false, showTotals)
			m.Width = 80
			m.Positions = positionMap
			m.Quotes = []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:    "AAPL",
						ShortName: "Apple Inc.",
					},
					Price:                   1.00 + change,
					Change:                  change,
					ChangePercent:           change,
					IsActive:                isActive,
					IsRegularTradingSession: isRegularTradingSession,
				},
			}
			Expect(removeFormatting(m.View())).To(Equal(expected))
		},
		Entry(
			describe("gain"),
			true,
			true,
			false,
			0.05,
			Position{},
			strings.Join([]string{
				"AAPL                       ⦿                                                1.05",
				"Apple Inc.                                                       ↑ 0.05  (0.05%)",
			}, "\n"),
		),
		Entry(
			describe("loss"),
			true,
			true,
			false,
			-0.05,
			Position{},
			strings.Join([]string{
				"AAPL                       ⦿                                                0.95",
				"Apple Inc.                                                      ↓ -0.05 (-0.05%)",
			}, "\n"),
		),
		Entry(
			describe("gain, after hours"),
			true,
			false,
			false,
			0.05,
			Position{},
			strings.Join([]string{
				"AAPL                       ⦾                                                1.05",
				"Apple Inc.                                                       ↑ 0.05  (0.05%)",
			}, "\n"),
		),
		Entry(
			describe("position, gain"),
			true,
			true,
			false,
			0.05,
			Position{
				AggregatedLot: AggregatedLot{
					Symbol:   "AAPL",
					Quantity: 100.0,
					Cost:     100.0,
				},
				Value:            105.0,
				DayChange:        5.0,
				DayChangePercent: 5.0,
			},
			strings.Join([]string{
				"AAPL                       ⦿                     105.00                     1.05",
				"Apple Inc.                              ↑ 5.00  (5.00%)          ↑ 0.05  (0.05%)",
			}, "\n"),
		),
		Entry(
			describe("position, loss"),
			true,
			true,
			false,
			-0.05,
			Position{
				AggregatedLot: AggregatedLot{
					Symbol:   "AAPL",
					Quantity: 100.0,
					Cost:     100.0,
				},
				Value:            95.0,
				DayChange:        -5.0,
				DayChangePercent: -5.0,
			},
			strings.Join([]string{
				"AAPL                       ⦿                      95.00                     0.95",
				"Apple Inc.                             ↓ -5.00 (-5.00%)         ↓ -0.05 (-0.05%)",
			}, "\n"),
		),
		Entry(
			describe("position, closed market"),
			false,
			false,
			false,
			0.0,
			Position{
				AggregatedLot: AggregatedLot{
					Symbol:   "AAPL",
					Quantity: 100.0,
					Cost:     100.0,
				},
				Value:            95.0,
				DayChange:        0.0,
				DayChangePercent: 0.0,
			},
			strings.Join([]string{
				"AAPL                                              95.00                     1.00",
				"Apple Inc.                                                         0.00  (0.00%)",
			}, "\n"),
		),
		Entry(
			describe("position, show totals"),
			false,
			false,
			true,
			0.0,
			Position{
				AggregatedLot: AggregatedLot{
					Symbol:   "AAPL",
					Quantity: 100.0,
					Cost:     100.0,
				},
				Value:            95.0,
				DayChange:        0.0,
				DayChangePercent: 0.0,
			},
			strings.Join([]string{
				"AAPL                                              95.00                     1.00",
				"Apple Inc.                                                         0.00  (0.00%)",
				"                                                                                ",
				"          Total Current Value:                    95.00                         ",
				"             Total Cost Basis:                   100.00                         ",
				"           Total Gains/Losses:         ↓ -5.00 (-0.05%)                         ",
			}, "\n"),
		),
	)

	When("there are more than one symbols on the watchlist", func() {
		It("should render a watchlist with each symbol", func() {

			m := NewModel(false, false, false, false)
			m.Width = 80
			m.Quotes = []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:    "AAPL",
						ShortName: "Apple Inc.",
					},
					Price:                   1.05,
					Change:                  0.00,
					ChangePercent:           0.00,
					IsActive:                false,
					IsRegularTradingSession: false,
				},
				{
					ResponseQuote: ResponseQuote{
						Symbol:    "TW",
						ShortName: "ThoughtWorks",
					},
					Price:                   109.04,
					Change:                  3.53,
					ChangePercent:           5.65,
					IsActive:                true,
					IsRegularTradingSession: false,
				},
				{
					ResponseQuote: ResponseQuote{
						Symbol:    "GOOG",
						ShortName: "Google Inc.",
					},
					Price:                   2523.53,
					Change:                  -32.02,
					ChangePercent:           -1.35,
					IsActive:                true,
					IsRegularTradingSession: false,
				},
				{
					ResponseQuote: ResponseQuote{
						Symbol:    "BTC-USD",
						ShortName: "Bitcoin",
					},
					Price:                   50000.0,
					Change:                  10000.0,
					ChangePercent:           20.0,
					IsActive:                true,
					IsRegularTradingSession: true,
				},
			}
			expected := strings.Join([]string{
				"BTC-USD                    ⦿                                            50000.00",
				"Bitcoin                                                     ↑ 10000.00  (20.00%)",
				"TW                         ⦾                                              109.04",
				"ThoughtWorks                                                     ↑ 3.53  (5.65%)",
				"GOOG                       ⦾                                             2523.53",
				"Google Inc.                                                    ↓ -32.02 (-1.35%)",
				"AAPL                                                                        1.05",
				"Apple Inc.                                                         0.00  (0.00%)",
			}, "\n")
			Expect(removeFormatting(m.View())).To(Equal(expected))
		})

		When("the show-separator layout flag is set", func() {
			It("should render a watchlist with separators", func() {

				m := NewModel(true, false, false, false)
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:    "BTC-USD",
							ShortName: "Bitcoin",
						},
						Price:                   50000.0,
						Change:                  10000.0,
						ChangePercent:           20.0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
					{
						ResponseQuote: ResponseQuote{
							Symbol:    "TW",
							ShortName: "ThoughtWorks",
						},
						Price:                   109.04,
						Change:                  3.53,
						ChangePercent:           5.65,
						IsActive:                true,
						IsRegularTradingSession: false,
					},
					{
						ResponseQuote: ResponseQuote{
							Symbol:    "GOOG",
							ShortName: "Google Inc.",
						},
						Price:                   2523.53,
						Change:                  -32.02,
						ChangePercent:           -1.35,
						IsActive:                true,
						IsRegularTradingSession: false,
					},
				}
				expected := strings.Join([]string{
					"BTC-USD                    ⦿                                            50000.00",
					"Bitcoin                                                     ↑ 10000.00  (20.00%)",
					"⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯",
					"TW                         ⦾                                              109.04",
					"ThoughtWorks                                                     ↑ 3.53  (5.65%)",
					"⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯⎯",
					"GOOG                       ⦾                                             2523.53",
					"Google Inc.                                                    ↓ -32.02 (-1.35%)",
				}, "\n")
				Expect(removeFormatting(m.View())).To(Equal(expected))
			})
		})
	})

	When("the option for extra exchange information is set", func() {
		It("should render extra exchange information", func() {
			m := NewModel(true, true, false, false)
			m.Quotes = []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:        "BTC-USD",
						ShortName:     "Bitcoin",
						Currency:      "USD",
						ExchangeName:  "Cryptocurrency",
						ExchangeDelay: 0,
					},
					Price:                   50000.0,
					Change:                  10000.0,
					ChangePercent:           20.0,
					IsActive:                true,
					IsRegularTradingSession: true,
				},
			}
			expected := strings.Join([]string{
				"BTC-USD                    ⦿                                            50000.00",
				"Bitcoin                                                     ↑ 10000.00  (20.00%)",
				" USD   Real-Time   Cryptocurrency                                               ",
			}, "\n")
			Expect(removeFormatting(m.View())).To(Equal(expected))
		})

		When("the exchange has a delay", func() {
			It("should render extra exchange information with the delay amount", func() {
				m := NewModel(true, true, false, false)
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:        "BTC-USD",
							ShortName:     "Bitcoin",
							Currency:      "USD",
							ExchangeName:  "Cryptocurrency",
							ExchangeDelay: 15,
						},
						Price:                   50000.0,
						Change:                  10000.0,
						ChangePercent:           20.0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
				}
				expected := strings.Join([]string{
					"BTC-USD                    ⦿                                            50000.00",
					"Bitcoin                                                     ↑ 10000.00  (20.00%)",
					" USD   Delayed 15min   Cryptocurrency                                           ",
				}, "\n")
				Expect(removeFormatting(m.View())).To(Equal(expected))
			})
		})
	})

	When("the option for extra fundamental information is set", func() {
		It("should render extra fundamental information", func() {
			m := NewModel(true, false, true, false)
			m.Quotes = []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "BTC-USD",
						ShortName:                  "Bitcoin",
						RegularMarketPreviousClose: 10000.0,
						RegularMarketOpen:          10000.0,
						RegularMarketDayRange:      "10000 - 10000",
					},
					Price:                   50000.0,
					Change:                  10000.0,
					ChangePercent:           20.0,
					IsActive:                true,
					IsRegularTradingSession: true,
				},
			}
			expected := strings.Join([]string{
				"BTC-USD                    ⦿                                            50000.00",
				"Bitcoin                                                     ↑ 10000.00  (20.00%)",
				"Prev Close: 10000.00     Open: 10000.00      Day Range: 10000 - 10000           ",
			}, "\n")
			Expect(removeFormatting(m.View())).To(Equal(expected))
		})

		When("there is no day range", func() {
			It("should not render the day range field", func() {
				m := NewModel(true, false, true, false)
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:                     "BTC-USD",
							ShortName:                  "Bitcoin",
							RegularMarketPreviousClose: 10000.0,
							RegularMarketOpen:          10000.0,
						},
						Price:                   50000.0,
						Change:                  10000.0,
						ChangePercent:           20.0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
				}
				expected := strings.Join([]string{
					"BTC-USD                    ⦿                                            50000.00",
					"Bitcoin                                                     ↑ 10000.00  (20.00%)",
					"Prev Close: 10000.00     Open: 10000.00                                         ",
				}, "\n")
				Expect(removeFormatting(m.View())).To(Equal(expected))
			})
		})
	})

	When("no quotes are set", func() {
		It("should render an empty watchlist", func() {
			m := NewModel(false, false, false, false)
			Expect(m.View()).To(Equal(""))
		})
	})

	When("the window width is less than the minimum", func() {
		It("should render an empty watchlist", func() {
			m := NewModel(false, false, false, false)
			m.Width = 70
			Expect(m.View()).To(Equal("Terminal window too narrow to render content\nResize to fix (70/80)"))
		})
	})
})

package watchlist_test

import (
	"io/ioutil"
	"strings"

	"github.com/acarl005/stripansi"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/internal/common"
	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"
	. "github.com/achannarasappa/ticker/internal/ui/component/watchlist"
)

func removeFormatting(text string) string {
	return stripansi.Strip(text)
}

func getLine(text string, lineIndex int) string {
	return strings.Split(text, "\n")[lineIndex]
}

var _ = Describe("Watchlist", func() {

	stylesFixture := c.Styles{
		Text:      func(v string) string { return v },
		TextLight: func(v string) string { return v },
		TextLabel: func(v string) string { return v },
		TextBold:  func(v string) string { return v },
		TextLine:  func(v string) string { return v },
		TextPrice: func(percent float64, text string) string { return text },
		Tag:       func(v string) string { return v },
	}

	It("should render a watchlist", func() {
		m := NewModel(c.Context{
			Reference: c.Reference{Styles: stylesFixture},
			Config: c.Config{
				ShowHoldings:          true,
				ExtraInfoExchange:     true,
				ExtraInfoFundamentals: true,
				Sort:                  "alpha",
			},
		})
		m.Width = 150
		m.Positions = map[string]Position{
			"STOCK4": {
				AggregatedLot: AggregatedLot{Symbol: "STOCK4", Quantity: 100.0, Cost: 50.0},
				Value:         105.0,
				DayChange:     5.0, DayChangePercent: 5.0,
				TotalChange: 55.0, TotalChangePercent: 110.0,
			},
			"STOCK5": {
				AggregatedLot: AggregatedLot{Symbol: "STOCK5", Quantity: 100.0, Cost: 150.0},
				Value:         105.0,
				DayChange:     5.0, DayChangePercent: 5.0,
				TotalChange: -45.0, TotalChangePercent: -30.0,
			},
			"STOCK6": {
				AggregatedLot: AggregatedLot{Symbol: "STOCK6", Quantity: 100.0, Cost: 50.0},
				Value:         95.0,
				DayChange:     -5.0, DayChangePercent: -5.0,
				TotalChange: -55.0, TotalChangePercent: -36.67,
			},
			"STOCK7": {
				AggregatedLot: AggregatedLot{Symbol: "STOCK7", Quantity: 100.0, Cost: 50.0},
				Value:         95.0,
				DayChange:     -5.0, DayChangePercent: -5.0,
				TotalChange: 45.0, TotalChangePercent: 90.0,
			},
			"STOCK8": {
				AggregatedLot: AggregatedLot{Symbol: "STOCK8", Quantity: 100.0, Cost: 100.0},
				Value:         95.0,
				DayChange:     0.0, DayChangePercent: 0.0,
				TotalChange: 45.0, TotalChangePercent: 90.0,
			},
		}
		m.Quotes = []Quote{
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK1", ShortName: "Stock 1 Inc. (gain)"},
				Price:         105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: 5.0, ChangePercent: 5.0,
				IsActive: true, IsRegularTradingSession: true,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK2", ShortName: "Stock 2 Inc. (loss)", FiftyTwoWeekHigh: 150.00},
				Price:         95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: -5.0, ChangePercent: -5.0,
				IsActive: true, IsRegularTradingSession: true,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK3", ShortName: "Stock 3 Inc. (gain, after hours)", FiftyTwoWeekHigh: 150.00},
				Price:         105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: 5.0, ChangePercent: 5.0,
				IsActive: true, IsRegularTradingSession: false,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK4", ShortName: "Stock 4 Inc. (position, day gain, total gain)", FiftyTwoWeekHigh: 150.00},
				Price:         105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: 5.0, ChangePercent: 5.0,
				IsActive: true, IsRegularTradingSession: true,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK5", ShortName: "Stock 5 Inc. (position, day gain, total loss)", FiftyTwoWeekHigh: 150.00},
				Price:         105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: 5.0, ChangePercent: 5.0,
				IsActive: true, IsRegularTradingSession: true,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK6", ShortName: "Stock 6 Inc. (position, day loss, total loss)", FiftyTwoWeekHigh: 150.00},
				Price:         95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: -5.0, ChangePercent: -5.0,
				IsActive: true, IsRegularTradingSession: true,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK7", ShortName: "Stock 7 Inc. (position, day loss, total gain)"},
				Price:         95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: -5.0, ChangePercent: -5.0,
				IsActive: true, IsRegularTradingSession: true,
			},
			{
				ResponseQuote: ResponseQuote{Symbol: "STOCK8", ShortName: "Stock 8 Inc. (position, closed market)"},
				Price:         95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00, PriceDayLow: 90.00,
				Change: 0.0, ChangePercent: 0.0,
				IsActive: false, IsRegularTradingSession: false,
			},
		}
		expected, _ := ioutil.ReadFile("./snapshots/watchlist-all-options.snap")
		Expect("\n" + removeFormatting(m.View())).To(BeIdenticalTo("\n" + string(expected)))
	})

	When("there are more than one symbols on the watchlist", func() {

		When("the show-separator layout flag is set", func() {
			It("should render a watchlist with separators", func() {

				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						Separate:              true,
						ExtraInfoExchange:     false,
						ExtraInfoFundamentals: false,
						Sort:                  "",
					},
				})
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

				expected := "────────────────────────────────────────────────────────────────────────────────"
				Expect("\n" + getLine(removeFormatting(m.View()), 2)).To(Equal("\n" + expected))
				Expect("\n" + getLine(removeFormatting(m.View()), 5)).To(Equal("\n" + expected))
				Expect("\n" + getLine(removeFormatting(m.View()), 8)).To(Equal("\n" + expected))
			})
		})
	})

	When("the option for extra exchange information is set", func() {
		It("should render extra exchange information", func() {
			m := NewModel(c.Context{
				Reference: c.Reference{Styles: stylesFixture},
				Config: c.Config{
					ExtraInfoExchange: true,
				},
			})
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
			expected := " USD   Real-Time   Cryptocurrency                                               "
			Expect("\n" + getLine(removeFormatting(m.View()), 2)).To(Equal("\n" + expected))
		})

		When("the exchange has a delay", func() {
			It("should render extra exchange information with the delay amount", func() {
				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						ExtraInfoExchange: true,
					},
				})
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
				expected := " USD   Delayed 15min   Cryptocurrency                                           "
				Expect("\n" + getLine(removeFormatting(m.View()), 2)).To(Equal("\n" + expected))
			})
		})

		When("the currency is being converted", func() {
			It("should show an indicator with the to and from currency codes", func() {
				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						ExtraInfoExchange: true,
						Currency:          "EUR",
					},
				})
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:        "APPL",
							ShortName:     "Apple, Inc",
							Currency:      "USD",
							ExchangeName:  "NASDAQ",
							ExchangeDelay: 0,
						},
						Price:                   5000.0,
						Change:                  1000.0,
						ChangePercent:           20.0,
						IsActive:                true,
						IsRegularTradingSession: true,
						CurrencyConverted:       "EUR",
					},
				}
				m.Context.Config.Currency = "EUR"
				expected := " USD → EUR   Real-Time   NASDAQ                                                 "
				Expect("\n" + getLine(removeFormatting(m.View()), 2)).To(Equal("\n" + expected))
			})

		})
	})

	When("the option for extra fundamental information is set", func() {
		It("should render extra fundamental information", func() {
			m := NewModel(c.Context{
				Reference: c.Reference{Styles: stylesFixture},
				Config: c.Config{
					ExtraInfoFundamentals: true,
				},
			})
			m.Width = 165
			m.Quotes = []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:                     "BTC-USD",
						ShortName:                  "Bitcoin",
						RegularMarketPreviousClose: 1000.0,
						RegularMarketOpen:          1000.0,
						RegularMarketDayHigh:       1500.0,
						RegularMarketDayLow:        500.0,
						FiftyTwoWeekHigh:           2000.0,
						FiftyTwoWeekLow:            300.0,
					},
					Price:                   5000.0,
					PricePrevClose:          1000.0,
					PriceOpen:               1000.0,
					PriceDayHigh:            200.0,
					PriceDayLow:             100.0,
					Change:                  1000.0,
					ChangePercent:           20.0,
					IsActive:                true,
					IsRegularTradingSession: true,
					IsVariablePrecision:     false,
				},
			}

			Expect(removeFormatting(m.View())).To(ContainSubstring("Day Range"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("52wk Range"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("100.00 - 200.00"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("300.00 - 2000.00"))
		})

		When("there is no day range", func() {
			It("should not render the day range field", func() {
				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						ExtraInfoFundamentals: true,
					},
				})
				m.Width = 135
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:                     "BTC-USD",
							ShortName:                  "Bitcoin",
							RegularMarketPreviousClose: 1000.0,
							RegularMarketOpen:          1000.0,
						},
						Price:                   5000.0,
						PricePrevClose:          1000.0,
						PriceOpen:               1000.0,
						Change:                  1000.0,
						ChangePercent:           20.0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
				}

				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Day Range"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("52wk Range"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("100.00 - 200.00"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("300.00 - 2000.00"))
			})
		})
	})

	When("the option for extra holding information is set", func() {
		It("should render extra holding information", func() {
			m := NewModel(c.Context{
				Reference: c.Reference{Styles: stylesFixture},
				Config: c.Config{
					ShowHoldings: true,
				},
			})
			m.Width = 120
			m.Quotes = []Quote{
				{
					ResponseQuote: ResponseQuote{
						Symbol:             "PTON",
						ShortName:          "Peloton",
						RegularMarketPrice: 100.0,
					},
					Price:                   100.0,
					Change:                  10.0,
					ChangePercent:           10.0,
					IsActive:                true,
					IsRegularTradingSession: true,
				},
			}
			m.Positions = map[string]Position{
				"PTON": {
					AggregatedLot: AggregatedLot{
						Symbol:   "PTON",
						Quantity: 100.0,
						Cost:     50.0,
					},
					Value:              105.0,
					DayChange:          5.0,
					DayChangePercent:   5.0,
					TotalChange:        55.0,
					TotalChangePercent: 110.0,
				},
			}
			Expect(removeFormatting(m.View())).To(ContainSubstring("Quantity"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("Avg. Cost"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("100.00"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("0.00"))
		})

		When("the holding quantity is high", func() {
			It("should render extra holding information without truncation", func() {
				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						ShowHoldings: true,
					},
				})
				m.Width = 120
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:             "PENNY",
							ShortName:          "A Penny Stock",
							RegularMarketPrice: 0.10,
						},
						Price:                   0.11,
						Change:                  0.01,
						ChangePercent:           10.0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
				}
				m.Positions = map[string]Position{
					"PENNY": {
						AggregatedLot: AggregatedLot{
							Symbol:   "PENNY",
							Quantity: 92709.0,
							Cost:     0.10,
						},
						Value:              9270.90,
						DayChange:          10.0,
						DayChangePercent:   10.0,
						TotalChange:        10.0,
						TotalChangePercent: 10.0,
					},
				}
				Expect(removeFormatting(m.View())).To(ContainSubstring("Quantity"))
				Expect(removeFormatting(m.View())).To(ContainSubstring("Avg. Cost"))
				Expect(removeFormatting(m.View())).To(ContainSubstring("92709.00"))
				Expect(removeFormatting(m.View())).To(ContainSubstring("0.00"))
			})
		})

		When("there is no position", func() {
			It("should not render quantity or average cost", func() {
				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						ShowHoldings: true,
					},
				})
				m.Width = 120
				m.Quotes = []Quote{
					{
						ResponseQuote: ResponseQuote{
							Symbol:             "PTON",
							ShortName:          "Peloton",
							RegularMarketPrice: 100.0,
						},
						Price:                   100.0,
						Change:                  10.0,
						ChangePercent:           10.0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
				}
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Quantity"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Avg. Cost"))
			})
		})
	})

	When("no quotes are set", func() {
		It("should render an empty watchlist", func() {
			m := NewModel(c.Context{
				Reference: c.Reference{Styles: stylesFixture},
				Config: c.Config{
					Separate:              false,
					ExtraInfoExchange:     false,
					ExtraInfoFundamentals: false,
					Sort:                  "",
				},
			})
			Expect(m.View()).To(Equal(""))
		})
	})

	When("the window width is less than the minimum", func() {
		It("should render an empty watchlist", func() {
			m := NewModel(c.Context{
				Reference: c.Reference{Styles: stylesFixture},
				Config: c.Config{
					Separate:              false,
					ExtraInfoExchange:     false,
					ExtraInfoFundamentals: false,
					Sort:                  "",
				},
			})
			m.Width = 70
			Expect(m.View()).To(Equal("Terminal window too narrow to render content\nResize to fix (70/80)"))
		})
	})
})

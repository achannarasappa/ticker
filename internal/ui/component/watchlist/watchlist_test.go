package watchlist_test

import (
	"io/ioutil"
	"strings"

	"github.com/acarl005/stripansi"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/ui/component/watchlist"
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
		m.Width = 175
		m.Assets = []c.Asset{
			{
				Symbol: "STOCK1", Name: "Stock 1 Inc. (gain)", QuoteExtended: c.QuoteExtended{MarketCap: 23467907, Volume: 4239786698},
				QuotePrice: c.QuotePrice{Price: 105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: 5.0, ChangePercent: 5.0},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: true},
			},
			{
				Symbol: "STOCK2", Name: "Stock 2 Inc. (loss)", QuoteExtended: c.QuoteExtended{FiftyTwoWeekHigh: 150.00},
				QuotePrice: c.QuotePrice{Price: 95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: -5.0, ChangePercent: -5.0},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: true},
			},
			{
				Symbol: "STOCK3", Name: "Stock 3 Inc. (gain, after hours)", QuoteExtended: c.QuoteExtended{FiftyTwoWeekHigh: 150.00},
				QuotePrice: c.QuotePrice{Price: 105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: 5.0, ChangePercent: 5.0},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: false},
			},
			{
				Symbol: "STOCK4", Name: "Stock 4 Inc. (position, day gain, total gain)", QuoteExtended: c.QuoteExtended{FiftyTwoWeekHigh: 150.00},
				QuotePrice: c.QuotePrice{Price: 105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: 5.0, ChangePercent: 5.0},
				Holding: c.Holding{
					Quantity:    100.0,
					Cost:        50.0,
					Value:       105.0,
					DayChange:   c.HoldingChange{Amount: 5.0, Percent: 5.0},
					TotalChange: c.HoldingChange{Amount: 55.0, Percent: 110.0},
				},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: true},
			},
			{
				Symbol: "STOCK5", Name: "Stock 5 Inc. (position, day gain, total loss)", QuoteExtended: c.QuoteExtended{FiftyTwoWeekHigh: 150.00},
				QuotePrice: c.QuotePrice{Price: 105.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: 5.0, ChangePercent: 5.0},
				Holding: c.Holding{
					Quantity:    100.0,
					Cost:        150.0,
					Value:       105.0,
					DayChange:   c.HoldingChange{Amount: 5.0, Percent: 5.0},
					TotalChange: c.HoldingChange{Amount: -45.0, Percent: -30.0},
				},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: true},
			},
			{
				Symbol: "STOCK6", Name: "Stock 6 Inc. (position, day loss, total loss)", QuoteExtended: c.QuoteExtended{FiftyTwoWeekHigh: 150.00},
				QuotePrice: c.QuotePrice{Price: 95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: -5.0, ChangePercent: -5.0},
				Holding: c.Holding{
					Quantity:    100.0,
					Cost:        50.0,
					Value:       95.0,
					DayChange:   c.HoldingChange{Amount: -5.0, Percent: -5.0},
					TotalChange: c.HoldingChange{Amount: -55.0, Percent: -36.67},
				},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: true},
			},
			{
				Symbol: "STOCK7", Name: "Stock 7 Inc. (position, day loss, total gain)",
				QuotePrice: c.QuotePrice{Price: 95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: -5.0, ChangePercent: -5.0},
				Holding: c.Holding{
					Quantity:    100.0,
					Cost:        50.0,
					Value:       95.0,
					DayChange:   c.HoldingChange{Amount: -5.0, Percent: -5.0},
					TotalChange: c.HoldingChange{Amount: 45.0, Percent: 90.0},
				},
				Exchange: c.Exchange{IsActive: true, IsRegularTradingSession: true},
			},
			{
				Symbol: "STOCK8", Name: "Stock 8 Inc. (position, closed market)",
				QuotePrice: c.QuotePrice{Price: 95.00, PricePrevClose: 100.00, PriceOpen: 110.00, PriceDayHigh: 120.00,
					PriceDayLow: 90.00, Change: 0.0, ChangePercent: 0.0},
				Holding: c.Holding{
					Quantity:    100.0,
					Cost:        100.0,
					Value:       95.0,
					DayChange:   c.HoldingChange{Amount: 0.0, Percent: 0.0},
					TotalChange: c.HoldingChange{Amount: 45.0, Percent: 90.0},
				},
				Exchange: c.Exchange{IsActive: false, IsRegularTradingSession: false},
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
				m.Assets = []c.Asset{
					{

						Symbol: "BTC-USD",
						Name:   "Bitcoin",
						QuotePrice: c.QuotePrice{
							Price:         50000.0,
							Change:        10000.0,
							ChangePercent: 20.0,
						},
						Exchange: c.Exchange{
							IsActive:                true,
							IsRegularTradingSession: true,
						},
					},
					{

						Symbol: "TW",
						Name:   "ThoughtWorks",
						QuotePrice: c.QuotePrice{
							Price:         109.04,
							Change:        3.53,
							ChangePercent: 5.65,
						},
						Exchange: c.Exchange{
							IsActive:                true,
							IsRegularTradingSession: false,
						},
					},
					{

						Symbol: "GOOG",
						Name:   "Google Inc.",
						QuotePrice: c.QuotePrice{
							Price:         2523.53,
							Change:        -32.02,
							ChangePercent: -1.35,
						},
						Exchange: c.Exchange{
							IsActive:                true,
							IsRegularTradingSession: false,
						},
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
			m.Assets = []c.Asset{
				{

					Symbol: "BTC-USD",
					Name:   "Bitcoin",
					Currency: c.Currency{
						FromCurrencyCode: "USD",
					},
					QuotePrice: c.QuotePrice{
						Price:         50000.0,
						Change:        10000.0,
						ChangePercent: 20.0,
					},
					Exchange: c.Exchange{
						Name:                    "Cryptocurrency",
						Delay:                   0,
						IsActive:                true,
						IsRegularTradingSession: true,
					},
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
				m.Assets = []c.Asset{
					{
						Symbol: "BTC-USD",
						Name:   "Bitcoin",
						Currency: c.Currency{
							FromCurrencyCode: "USD",
						},
						QuotePrice: c.QuotePrice{
							Price:         50000.0,
							Change:        10000.0,
							ChangePercent: 20.0,
						},
						Exchange: c.Exchange{
							Name:                    "Cryptocurrency",
							Delay:                   15,
							IsActive:                true,
							IsRegularTradingSession: true,
						},
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
				m.Assets = []c.Asset{
					{
						Symbol: "APPL",
						Name:   "Apple, Inc",
						Currency: c.Currency{
							FromCurrencyCode: "USD",
							ToCurrencyCode:   "EUR",
						},
						QuotePrice: c.QuotePrice{
							Price:         5000.0,
							Change:        1000.0,
							ChangePercent: 20.0,
						},
						Exchange: c.Exchange{
							Name:                    "NASDAQ",
							Delay:                   0,
							IsActive:                true,
							IsRegularTradingSession: true,
						},
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
			m.Assets = []c.Asset{
				{
					Symbol: "BTC-USD",
					Name:   "Bitcoin",
					QuotePrice: c.QuotePrice{
						Price:          5000.0,
						PricePrevClose: 1000.0,
						PriceOpen:      1000.0,
						PriceDayHigh:   200.0,
						PriceDayLow:    100.0,
						Change:         1000.0,
						ChangePercent:  20.0,
					},
					QuoteExtended: c.QuoteExtended{
						FiftyTwoWeekHigh: 2000.0,
						FiftyTwoWeekLow:  300.0,
					},
					Exchange: c.Exchange{
						IsActive:                true,
						IsRegularTradingSession: true,
					},
					Meta: c.Meta{
						IsVariablePrecision: false,
					},
				},
			}

			Expect(removeFormatting(m.View())).To(ContainSubstring("Day Range"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("52wk Range"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("100.00 - 200.00"))
			Expect(removeFormatting(m.View())).To(ContainSubstring("300.00 - 2000.00"))
		})

		When("there is no day range or open price", func() {

			m := NewModel(c.Context{
				Reference: c.Reference{Styles: stylesFixture},
				Config: c.Config{
					ExtraInfoFundamentals: true,
				},
			})
			m.Width = 135
			m.Assets = []c.Asset{
				{
					Symbol: "BTC-USD",
					Name:   "Bitcoin",
					QuotePrice: c.QuotePrice{
						Price:          5000.0,
						PricePrevClose: 1000.0,
						PriceOpen:      0.0,
						Change:         1000.0,
						ChangePercent:  20.0,
					},
					Exchange: c.Exchange{
						IsActive:                true,
						IsRegularTradingSession: true,
					},
				},
			}

			It("should not render the day range field", func() {
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Day Range"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("52wk Range"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("100.00 - 200.00"))
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("300.00 - 2000.00"))
			})

			It("should render a placeholder for the open price", func() {
				Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Open:"))
			})

		})

		When("the asset is a futures contract", func() {
			It("should render the underlying asset symbol", func() {
				m := NewModel(c.Context{
					Reference: c.Reference{Styles: stylesFixture},
					Config: c.Config{
						ExtraInfoFundamentals: true,
					},
				})
				m.Width = 150
				m.Assets = []c.Asset{
					{
						Symbol: "BIT-27DEC24-CDE",
						Name:   "Nano Bitcoin Futures",
						Class:  c.AssetClassFuturesContract,
						QuotePrice: c.QuotePrice{
							Price:          50333.0,
							PricePrevClose: 1000.0,
							PriceOpen:      0.0,
							Change:         1000.0,
							ChangePercent:  20.0,
						},
						Exchange: c.Exchange{
							IsActive:                true,
							IsRegularTradingSession: true,
						},
						QuoteFutures: c.QuoteFutures{
							IndexPrice: 50312,
							Basis:      10.0,
							Expiry:     "5d 10h",
						},
					},
				}
				Expect(removeFormatting(m.View())).To(ContainSubstring("50312"))
				Expect(removeFormatting(m.View())).To(ContainSubstring("5d 10h"))
				Expect(removeFormatting(m.View())).To(ContainSubstring("10.00%"))
			})

			When("the index price is not set", func() {
				It("should not render the index price", func() {
					m := NewModel(c.Context{
						Reference: c.Reference{Styles: stylesFixture},
						Config: c.Config{
							ExtraInfoFundamentals: true,
						},
					})
					m.Width = 150
					m.Assets = []c.Asset{
						{
							Symbol: "BIT-27DEC24-CDE",
							Name:   "Nano Bitcoin Futures",
							Class:  c.AssetClassFuturesContract,
							QuotePrice: c.QuotePrice{
								Price:          50333.0,
								PricePrevClose: 1000.0,
								PriceOpen:      0.0,
								Change:         1000.0,
								ChangePercent:  20.0,
							},
						},
					}
					Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Index Price"))
					Expect(removeFormatting(m.View())).ToNot(ContainSubstring("Basis"))
				})
			})

			When("the day range and expiry are set", func() {
				It("should render both fields", func() {
					m := NewModel(c.Context{
						Reference: c.Reference{Styles: stylesFixture},
						Config: c.Config{
							ExtraInfoFundamentals: true,
						},
					})
					m.Width = 150
					m.Assets = []c.Asset{
						{
							Symbol: "BIT-27DEC24-CDE",
							Name:   "Nano Bitcoin Futures",
							Class:  c.AssetClassFuturesContract,
							QuotePrice: c.QuotePrice{
								Price:         50333.0,
								PriceDayHigh:  50500.0,
								PriceDayLow:   49000.0,
								Change:        1000.0,
								ChangePercent: 20.0,
							},
							QuoteFutures: c.QuoteFutures{
								IndexPrice: 50312,
								Basis:      10.0,
								Expiry:     "5d 10h",
							},
						},
					}
					Expect(removeFormatting(m.View())).To(ContainSubstring("Day Range"))
					Expect(removeFormatting(m.View())).To(ContainSubstring("49000.00 - 50500.00"))
					Expect(removeFormatting(m.View())).To(ContainSubstring("Expiry"))
					Expect(removeFormatting(m.View())).To(ContainSubstring("5d 10h"))
				})
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
			m.Assets = []c.Asset{
				{
					Symbol: "PTON",
					Name:   "Peloton",
					QuotePrice: c.QuotePrice{
						Price:         100.0,
						Change:        10.0,
						ChangePercent: 10.0,
					},
					Holding: c.Holding{
						Quantity:    100.0,
						Cost:        50.0,
						Value:       105.0,
						DayChange:   c.HoldingChange{Amount: 5.0, Percent: 5.0},
						TotalChange: c.HoldingChange{Amount: 55.0, Percent: 110.0},
					},
					Exchange: c.Exchange{
						IsActive:                true,
						IsRegularTradingSession: true,
					},
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
				m.Assets = []c.Asset{
					{
						Symbol: "PENNY",
						Name:   "A Penny Stock",
						QuotePrice: c.QuotePrice{
							Price:         0.11,
							Change:        0.01,
							ChangePercent: 10.0,
						},
						Holding: c.Holding{
							Quantity:    92709.0,
							Cost:        0.10,
							Value:       9270.90,
							DayChange:   c.HoldingChange{Amount: 10.0, Percent: 10.0},
							TotalChange: c.HoldingChange{Amount: 10.0, Percent: 10.0},
						},
						Exchange: c.Exchange{
							IsActive:                true,
							IsRegularTradingSession: true,
						},
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
				m.Assets = []c.Asset{
					{
						Symbol: "PTON",
						Name:   "Peloton",
						QuotePrice: c.QuotePrice{
							Price:         100.0,
							Change:        10.0,
							ChangePercent: 10.0,
						},
						Exchange: c.Exchange{
							IsActive:                true,
							IsRegularTradingSession: true,
						},
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

package sorter_test

import (
	c "github.com/achannarasappa/ticker/v4/internal/common"
	. "github.com/achannarasappa/ticker/v4/internal/sorter"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorter", func() {

	Describe("NewSorter", func() {
		bitcoinQuote := c.Asset{
			Symbol: "BTC-USD",
			Name:   "Bitcoin",
			QuotePrice: c.QuotePrice{
				PricePrevClose: 10000.0,
				PriceOpen:      10000.0,
				Price:          50000.0,
				Change:         10000.0,
				ChangePercent:  20.0,
			},
			Holding: c.Holding{
				Value: 50000.0,
			},
			Exchange: c.Exchange{
				IsActive:                true,
				IsRegularTradingSession: true,
			},
			Meta: c.Meta{
				OrderIndex: 1,
			},
		}
		twQuote := c.Asset{
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
		}
		googleQuote := c.Asset{
			Symbol: "GOOG",
			Name:   "Google Inc.",
			QuotePrice: c.QuotePrice{
				Price:         2523.53,
				Change:        -32.02,
				ChangePercent: -1.35,
			},
			Holding: c.Holding{
				Value: 2523.53,
			},
			Exchange: c.Exchange{
				IsActive:                true,
				IsRegularTradingSession: false,
			},
			Meta: c.Meta{
				OrderIndex: 0,
			},
		}
		msftQuote := c.Asset{
			Symbol: "MSFT",
			Name:   "Microsoft Corporation",
			QuotePrice: c.QuotePrice{
				Price:         242.01,
				Change:        -0.99,
				ChangePercent: -0.41,
			},
			Exchange: c.Exchange{
				IsActive:                false,
				IsRegularTradingSession: false,
			},
		}
		rblxQuote := c.Asset{
			Symbol: "RBLX",
			Name:   "Roblox",
			QuotePrice: c.QuotePrice{
				Price:         85.00,
				Change:        10.00,
				ChangePercent: 7.32,
			},
			Exchange: c.Exchange{
				IsActive:                false,
				IsRegularTradingSession: false,
			},
		}
		assets := []c.Asset{
			bitcoinQuote,
			twQuote,
			googleQuote,
			msftQuote,
		}

		When("providing no sort parameter", func() {
			It("should sort by default (change percent)", func() {
				sorter := NewSorter("")

				coinQuote := c.Asset{
					Symbol: "COIN",
					Name:   "Coinbase",
					QuotePrice: c.QuotePrice{
						Price:         220.00,
						Change:        20.00,
						ChangePercent: 10.00,
					},
					Exchange: c.Exchange{
						IsActive:                false,
						IsRegularTradingSession: false,
					},
				}
				assets := []c.Asset{
					rblxQuote,
					bitcoinQuote,
					twQuote,
					googleQuote,
					msftQuote,
					coinQuote,
				}

				sortedQuotes := sorter(assets)
				expected := []c.Asset{
					bitcoinQuote,
					twQuote,
					googleQuote,
					coinQuote,
					rblxQuote,
					msftQuote,
				}

				Expect(sortedQuotes).To(Equal(expected))
			})
		})
		When("providing \"alpha\" as a sort parameter", func() {
			It("should sort by alphabetical order", func() {
				sorter := NewSorter("alpha")

				sortedQuotes := sorter(assets)
				expected := []c.Asset{
					bitcoinQuote,
					googleQuote,
					msftQuote,
					twQuote,
				}

				Expect(sortedQuotes).To(Equal(expected))
			})
		})
		When("providing \"position\" as a sort parameter", func() {
			It("should sort position value, with inactive quotes last", func() {
				sorter := NewSorter("value")

				bitcoinQuoteWithHolding := bitcoinQuote
				bitcoinQuoteWithHolding.Holding.Value = 50000.0
				googleQuoteWithHolding := googleQuote
				googleQuoteWithHolding.Holding.Value = 2523.53
				rblxQuoteWithHolding := rblxQuote
				rblxQuoteWithHolding.Holding.Value = 900.00
				msftQuoteWithHolding := msftQuote
				msftQuoteWithHolding.Holding.Value = 100.00

				assets := []c.Asset{
					bitcoinQuoteWithHolding,
					twQuote,
					googleQuoteWithHolding,
					msftQuoteWithHolding,
					rblxQuoteWithHolding,
				}

				sortedQuotes := sorter(assets)
				expected := []c.Asset{
					bitcoinQuoteWithHolding,
					googleQuoteWithHolding,
					twQuote,
					rblxQuoteWithHolding,
					msftQuoteWithHolding,
				}

				Expect(sortedQuotes).To(Equal(expected))
			})
		})
		When("providing \"user\" as a sort parameter", func() {
			It("should sort by the user defined order for positions and watchlist", func() {
				sorter := NewSorter("user")

				sortedQuotes := sorter(assets)
				expected := []c.Asset{
					googleQuote,
					bitcoinQuote,
					twQuote,
					msftQuote,
				}

				Expect(sortedQuotes).To(Equal(expected))
			})
		})
		When("providing no quotes", func() {
			When("default sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("")

					sortedQuotes := sorter([]c.Asset{})
					expected := []c.Asset{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
			When("alpha sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("alpha")

					sortedQuotes := sorter([]c.Asset{})
					expected := []c.Asset{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
			When("value sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("value")

					sortedQuotes := sorter([]c.Asset{})
					expected := []c.Asset{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
			When("user sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("user")

					sortedQuotes := sorter([]c.Asset{})
					expected := []c.Asset{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
		})
	})
})

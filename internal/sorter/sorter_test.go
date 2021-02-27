package sorter_test

import (
	. "github.com/achannarasappa/ticker/internal/position"
	. "github.com/achannarasappa/ticker/internal/quote"
	. "github.com/achannarasappa/ticker/internal/sorter"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Sorter", func() {

	Describe("NewSorter", func() {
		bitcoinQuote := Quote{
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
		}
		twQuote := Quote{
			ResponseQuote: ResponseQuote{
				Symbol:    "TW",
				ShortName: "ThoughtWorks",
			},
			Price:                   109.04,
			Change:                  3.53,
			ChangePercent:           5.65,
			IsActive:                true,
			IsRegularTradingSession: false,
		}
		googleQuote := Quote{
			ResponseQuote: ResponseQuote{
				Symbol:    "GOOG",
				ShortName: "Google Inc.",
			},
			Price:                   2523.53,
			Change:                  -32.02,
			ChangePercent:           -1.35,
			IsActive:                true,
			IsRegularTradingSession: false,
		}
		msftQuote := Quote{
			ResponseQuote: ResponseQuote{
				Symbol:    "MSFT",
				ShortName: "Microsoft Corporation",
			},
			Price:                   242.01,
			Change:                  -0.99,
			ChangePercent:           -0.41,
			IsActive:                false,
			IsRegularTradingSession: false,
		}
		quotes := []Quote{
			bitcoinQuote,
			twQuote,
			googleQuote,
			msftQuote,
		}

		positions := map[string]Position{
			"BTC-USD": {
				Value: 50000.0,
				AggregatedLot: AggregatedLot{
					OrderIndex: 1,
				},
			},
			"GOOG": {
				Value: 2523.53,
				AggregatedLot: AggregatedLot{
					OrderIndex: 0,
				},
			},
		}
		When("providing no sort parameter", func() {
			It("should sort by default (change percent)", func() {
				sorter := NewSorter("")

				sortedQuotes := sorter(quotes, positions)
				expected := []Quote{
					bitcoinQuote,
					twQuote,
					googleQuote,
					msftQuote,
				}

				Expect(sortedQuotes).To(Equal(expected))
			})
		})
		When("providing \"alpha\" as a sort parameter", func() {
			It("should sort by alphabetical order", func() {
				sorter := NewSorter("alpha")

				sortedQuotes := sorter(quotes, positions)
				expected := []Quote{
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

				sortedQuotes := sorter(quotes, positions)
				expected := []Quote{
					bitcoinQuote,
					googleQuote,
					twQuote,
					msftQuote,
				}

				Expect(sortedQuotes).To(Equal(expected))
			})
		})
		When("providing \"user\" as a sort parameter", func() {
			It("should sort by the user defined order for positions and watchlist", func() {
				sorter := NewSorter("user")

				sortedQuotes := sorter(quotes, positions)
				expected := []Quote{
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

					sortedQuotes := sorter([]Quote{}, map[string]Position{})
					expected := []Quote{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
			When("alpha sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("alpha")

					sortedQuotes := sorter([]Quote{}, map[string]Position{})
					expected := []Quote{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
			When("value sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("value")

					sortedQuotes := sorter([]Quote{}, map[string]Position{})
					expected := []Quote{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
			When("user sorter", func() {
				It("should return no quotes", func() {
					sorter := NewSorter("user")

					sortedQuotes := sorter([]Quote{}, map[string]Position{})
					expected := []Quote{}
					Expect(sortedQuotes).To(Equal(expected))
				})
			})
		})
	})
})

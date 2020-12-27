package main

import (
	"log"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// fmt.Printf("%#v \n", quote.GetQuotes([]string{"ABNB", "PLTR"}))

	// quotes := []quote.Quote{
	// 	{Symbol: "AAPL", ShortName: "Apple, Inc.", RegularMarketPrice: 1000.1, RegularMarketChange: 10.1, RegularMarketChangePercent: 1.1},
	// 	{Symbol: "ABNB", ShortName: "AirBnB, Inc.", RegularMarketPrice: 645.1, RegularMarketChange: 4.1, RegularMarketChangePercent: 0.9},
	// }

	symbols := []string{
		"NET",
		"TSLA",
		"MSFT",
		"OKTA",
		"TEAM",
		"GOOG",
		"DASH",
		"DIS",
		"NFLX",
		"FB",
		"AMZN",
		"ESTC",
		"ARKW",
		"BTC-USD",
	}

	p := tea.NewProgram(ui.NewModel(symbols, quote.GetQuotes))

	p.EnterAltScreen()
	err := p.Start()
	p.ExitAltScreen()

	if err != nil {
		log.Fatal(err)
	}

}

package main

import (
	"log"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

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

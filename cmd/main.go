package main

import (
	"log"
	"ticker-tape/internal/position"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/novalagung/gubrak/v2"
)

func main() {

	symbolsWatchlist := []string{
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
		"BTC-USD",
		"XRP-USD",
		"CMPS",
		"SNAP",
		"SNOW",
	}

	symbols := (gubrak.
		From(position.GetSymbols()).
		Concat(symbolsWatchlist).
		Uniq().
		Result()).([]string)

	positions := position.GetPositions(quote.GetQuotes(symbols))

	p := tea.NewProgram(ui.NewModel(symbols, positions, quote.GetQuotes))

	p.EnableMouseCellMotion()
	p.EnterAltScreen()
	err := p.Start()
	p.ExitAltScreen()
	p.DisableMouseCellMotion()

	if err != nil {
		log.Fatal(err)
	}

}

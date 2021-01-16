package main

import (
	"log"
	"os"
	"ticker-tape/internal/position"
	"ticker-tape/internal/quote"
	"ticker-tape/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-resty/resty/v2"
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

	var (
		err    error
		handle *os.File
	)

	handle, err = os.Open("./positions.yaml")

	if err != nil {
		log.Fatal(err)
	}
	defer handle.Close()

	lots := position.GetLots(handle)
	symbols := position.GetSymbols(symbolsWatchlist, lots)

	client := resty.New()

	p := tea.NewProgram(ui.NewModel(position.GetPositions(lots), quote.GetQuotes(*client, symbols)))

	p.EnableMouseCellMotion()
	p.EnterAltScreen()
	err = p.Start()
	p.ExitAltScreen()
	p.DisableMouseCellMotion()

	if err != nil {
		log.Fatal(err)
	}

}

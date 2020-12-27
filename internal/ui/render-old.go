package ui

// import (
// 	"log"
// 	"strconv"
// 	"ticker-tape/internal/quote"

// 	ui "github.com/gizak/termui/v3"
// 	"github.com/gizak/termui/v3/widgets"

// 	. "github.com/novalagung/gubrak"
// )

// func convertQuotesToRows(quotes []quote.Quote) [][]string {

// 	rows, _ := Reduce(quotes, func(acc [][]string, quote quote.Quote) [][]string {
// 		return append(acc, []string{
// 			quote.ShortName,
// 			strconv.FormatFloat(quote.RegularMarketPrice, 'f', 2, 64),
// 			strconv.FormatFloat(quote.RegularMarketChange, 'f', 2, 64),
// 			strconv.FormatFloat(quote.RegularMarketChangePercent, 'f', 2, 64),
// 		})
// 	}, [][]string{{"Symbol", "RegularMarketPrice", "RegularMarketChange", "RegularMarketChangePercent"}})

// 	return (rows).([][]string)

// }

// func renderQuotes(quotes []quote.Quote) {

// 	table1 := widgets.NewTable()
// 	table1.Rows = convertQuotesToRows(quotes)
// 	table1.TextStyle = ui.NewStyle(ui.ColorWhite)
// 	table1.BorderStyle = ui.NewStyle(ui.ColorBlack)
// 	table1.Border = false
// 	table1.RowStyles[0] = ui.NewStyle(ui.ColorCyan)
// 	table1.SetRect(0, 0, 50, 10)

// 	ui.Render(table1)

// }

// func handleInput() {
// 	uiEvents := ui.PollEvents()
// 	for {
// 		select {
// 		case e := <-uiEvents:
// 			switch e.ID {
// 			case "q", "<C-c>":
// 				ui.Close()
// 				return
// 			}
// 		}
// 	}
// }

// func Render(quotes []quote.Quote) {
// 	if err := ui.Init(); err != nil {
// 		log.Fatalf("failed to initialize termui: %v", err)
// 	}

// 	renderQuotes(quotes)
// 	handleInput()

// }

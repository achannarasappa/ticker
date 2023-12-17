package symbol

import (
	"encoding/csv"
	"errors"
	"io"

	c "github.com/achannarasappa/ticker/v4/internal/common"
	"github.com/go-resty/resty/v2"
)

type SymbolSourceMap struct { //nolint:golint,revive
	TickerSymbol string
	SourceSymbol string
	Source       c.QuoteSource
}

type TickerSymbolToSourceSymbol map[string]SymbolSourceMap

func parseQuoteSource(id string) c.QuoteSource {
	if id == "cg" {
		return c.QuoteSourceCoingecko
	}

	return c.QuoteSourceUnknown
}

func parseTickerSymbolToSourceSymbol(body io.ReadCloser) (TickerSymbolToSourceSymbol, error) {

	out := TickerSymbolToSourceSymbol{}
	reader := csv.NewReader(body)
	reader.LazyQuotes = true
	for {

		row, err := reader.Read()

		if errors.Is(err, io.EOF) {
			body.Close()

			break
		}

		if err != nil {
			return nil, err
		}

		if _, exists := out[row[0]]; !exists {
			out[row[0]] = SymbolSourceMap{
				TickerSymbol: row[0],
				SourceSymbol: row[1],
				Source:       parseQuoteSource(row[2]),
			}

		}
	}

	return out, nil
}

// GetTickerSymbols retrieves a list of ticker specific symbols and their data source
func GetTickerSymbols(client resty.Client) (TickerSymbolToSourceSymbol, error) {
	url := "https://raw.githubusercontent.com/achannarasappa/ticker-static/master/symbols.csv"
	res, err := client.R().
		SetDoNotParseResponse(true).
		Get(url)
	body := res.RawBody()

	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}

	tickerSymbolToSourceSymbol, err := parseTickerSymbolToSourceSymbol(body)

	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}

	return tickerSymbolToSourceSymbol, nil
}

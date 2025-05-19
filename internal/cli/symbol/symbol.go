package symbol

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"

	c "github.com/achannarasappa/ticker/v4/internal/common"
)

type SymbolSourceMap struct { //nolint:golint,revive
	TickerSymbol string
	SourceSymbol string
	Source       c.QuoteSource
}

type TickerSymbolToSourceSymbol map[string]SymbolSourceMap

func parseQuoteSource(id string) c.QuoteSource {

	if id == "cb" {
		return c.QuoteSourceCoinbase
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
func GetTickerSymbols(url string) (TickerSymbolToSourceSymbol, error) {
	resp, err := http.Get(url) //nolint:gosec
	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return TickerSymbolToSourceSymbol{}, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	tickerSymbolToSourceSymbol, err := parseTickerSymbolToSourceSymbol(resp.Body)
	if err != nil {
		return TickerSymbolToSourceSymbol{}, err
	}

	return tickerSymbolToSourceSymbol, nil
}

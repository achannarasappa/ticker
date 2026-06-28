package symbol

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// ttlSymbolMap is how long the symbol source map is cached. It is sourced from a
// static CSV that changes rarely, so it can be reused for a long time.
const ttlSymbolMap = 7 * 24 * time.Hour

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

// GetTickerSymbols retrieves a list of ticker specific symbols and their data
// source. When a cache is provided and holds a fresh entry, the symbols are
// served from it without a network request.
func GetTickerSymbols(url string, cache c.Cache) (TickerSymbolToSourceSymbol, error) {
	cacheKey := "symbols:" + url

	if cache != nil {
		var cached TickerSymbolToSourceSymbol
		if cache.Get(cacheKey, &cached) {
			return cached, nil
		}
	}

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

	if cache != nil {
		cache.Set(cacheKey, tickerSymbolToSourceSymbol, ttlSymbolMap)
	}

	return tickerSymbolToSourceSymbol, nil
}

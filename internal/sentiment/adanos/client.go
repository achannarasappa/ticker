package adanos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	common "github.com/achannarasappa/ticker/v5/internal/common"
)

var sourceIDs = []string{"reddit", "x", "news", "polymarket"} //nolint:gochecknoglobals

type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	nowFn      func() time.Time
	ttl        time.Duration
	mu         sync.RWMutex
	cache      map[string]cacheEntry
}

type cacheEntry struct {
	snapshot  common.MarketSentiment
	expiresAt time.Time
}

type sourceSnapshot struct {
	buzzScore      float64
	bullishPercent float64
	activityValue  float64
	available      bool
}

func NewClient(baseURL, apiKey string, httpClient *http.Client, ttl time.Duration) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	if ttl <= 0 {
		ttl = 5 * time.Minute
	}

	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		apiKey:     strings.TrimSpace(apiKey),
		httpClient: httpClient,
		nowFn:      time.Now,
		ttl:        ttl,
		cache:      make(map[string]cacheEntry),
	}
}

func (client *Client) Enabled() bool {
	return client != nil && client.apiKey != ""
}

func (client *Client) FetchSnapshots(ctx context.Context, symbols []string) (map[string]common.MarketSentiment, error) {
	results := make(map[string]common.MarketSentiment)
	if !client.Enabled() {
		return results, nil
	}

	uniqueSymbols := normalizeSymbols(symbols)
	missing := make([]string, 0, len(uniqueSymbols))

	client.mu.RLock()
	now := client.nowFn()
	for _, symbol := range uniqueSymbols {
		entry, ok := client.cache[symbol]
		if ok && now.Before(entry.expiresAt) {
			results[symbol] = entry.snapshot
			continue
		}
		missing = append(missing, symbol)
	}
	client.mu.RUnlock()

	if len(missing) == 0 {
		return results, nil
	}

	sourceValuesBySymbol := make(map[string][]sourceSnapshot, len(missing))
	var firstErr error
	for _, sourceID := range sourceIDs {
		sourceResults, err := client.fetchSource(ctx, sourceID, missing)
		if err != nil && firstErr == nil {
			firstErr = err
		}

		for _, symbol := range missing {
			if snapshot, ok := sourceResults[symbol]; ok && snapshot.available {
				sourceValuesBySymbol[symbol] = append(sourceValuesBySymbol[symbol], snapshot)
			}
		}
	}

	client.mu.Lock()
	defer client.mu.Unlock()
	expiresAt := client.nowFn().Add(client.ttl)
	for _, symbol := range missing {
		snapshot := aggregateSnapshots(sourceValuesBySymbol[symbol])
		if snapshot.Available {
			client.cache[symbol] = cacheEntry{snapshot: snapshot, expiresAt: expiresAt}
			results[symbol] = snapshot
		}
	}

	return results, firstErr
}

func (client *Client) fetchSource(ctx context.Context, sourceID string, symbols []string) (map[string]sourceSnapshot, error) {
	queryURL, err := url.Parse(client.baseURL + "/" + sourceID + "/stocks/v1/compare")
	if err != nil {
		return nil, err
	}

	values := queryURL.Query()
	values.Set("tickers", strings.Join(symbols, ","))
	values.Set("days", "7")
	queryURL.RawQuery = values.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, queryURL.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", client.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("adanos %s compare returned status %d", sourceID, resp.StatusCode) //nolint:goerr113
	}

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}

	return parseComparePayload(payload), nil
}

func normalizeSymbols(symbols []string) []string {
	seen := make(map[string]struct{}, len(symbols))
	uniqueSymbols := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		normalized := strings.ToUpper(strings.TrimSpace(symbol))
		if normalized == "" {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		uniqueSymbols = append(uniqueSymbols, normalized)
	}

	return uniqueSymbols
}

func parseComparePayload(payload map[string]any) map[string]sourceSnapshot {
	if data, ok := payload["data"].(map[string]any); ok {
		payload = data
	}

	stocks, _ := payload["stocks"].([]any)
	results := make(map[string]sourceSnapshot, len(stocks))
	for _, item := range stocks {
		stock, ok := item.(map[string]any)
		if !ok {
			continue
		}
		symbol := stringValue(stock, "ticker", "symbol")
		if symbol == "" {
			continue
		}
		results[symbol] = sourceSnapshot{
			buzzScore:      floatValue(stock, "buzz_score"),
			bullishPercent: floatValue(stock, "bullish_pct"),
			activityValue:  floatValue(stock, "mentions", "trade_count", "tradeCount"),
		}
		results[symbol] = markAvailability(results[symbol])
	}

	return results
}

func aggregateSnapshots(snapshots []sourceSnapshot) common.MarketSentiment {
	if len(snapshots) == 0 {
		return common.MarketSentiment{}
	}

	var totalBuzz float64
	var totalBullish float64
	bullishValues := make([]float64, 0, len(snapshots))
	for _, snapshot := range snapshots {
		totalBuzz += snapshot.buzzScore
		totalBullish += snapshot.bullishPercent
		bullishValues = append(bullishValues, snapshot.bullishPercent)
	}

	return common.MarketSentiment{
		Available:       true,
		AverageBuzz:     totalBuzz / float64(len(snapshots)),
		BullishPercent:  totalBullish / float64(len(snapshots)),
		Coverage:        len(snapshots),
		SourceAlignment: sourceAlignment(bullishValues),
	}
}

func sourceAlignment(values []float64) string {
	if len(values) == 0 {
		return "unavailable"
	}
	if len(values) == 1 {
		return "single_source"
	}

	min := values[0]
	max := values[0]
	for _, value := range values[1:] {
		if value < min {
			min = value
		}
		if value > max {
			max = value
		}
	}

	spread := max - min
	if spread <= 12 {
		return "aligned"
	}
	if spread <= 25 {
		return "mixed"
	}

	return "divergent"
}

func floatValue(values map[string]any, keys ...string) float64 {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case float64:
			return typed
		case int:
			return float64(typed)
		case string:
			parsed, _ := strconv.ParseFloat(typed, 64)
			return parsed
		}
	}

	return 0
}

func stringValue(values map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := values[key]
		if !ok {
			continue
		}
		if typed, ok := value.(string); ok {
			return strings.ToUpper(strings.TrimSpace(typed))
		}
	}
	return ""
}

func markAvailability(snapshot sourceSnapshot) sourceSnapshot {
	if snapshot.buzzScore > 0 || snapshot.activityValue > 0 {
		snapshot.available = true
	}

	return snapshot
}

package adanos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchSnapshotsAggregatesAcrossSourcesAndCaches(t *testing.T) {
	t.Helper()

	requestsByPath := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsByPath[r.URL.Path]++

		w.Header().Set("Content-Type", "application/json")

		var payload map[string]any
		switch r.URL.Path {
		case "/reddit/stocks/v1/compare":
			payload = map[string]any{
				"stocks": []map[string]any{
					{"ticker": "AAPL", "buzz_score": 40.0, "bullish_pct": 70.0, "mentions": 12.0},
				},
			}
		case "/x/stocks/v1/compare":
			payload = map[string]any{
				"data": map[string]any{
					"stocks": []map[string]any{
						{"ticker": "AAPL", "buzz_score": 30.0, "bullish_pct": 65.0, "mentions": 8.0},
					},
				},
			}
		case "/news/stocks/v1/compare":
			payload = map[string]any{
				"stocks": []map[string]any{
					{"ticker": "AAPL", "buzz_score": 0.0, "bullish_pct": 0.0, "mentions": 0.0},
				},
			}
		case "/polymarket/stocks/v1/compare":
			payload = map[string]any{
				"stocks": []map[string]any{
					{"ticker": "AAPL", "buzz_score": 25.0, "bullish_pct": 55.0, "trade_count": 4.0},
				},
			}
		default:
			http.NotFound(w, r)
			return
		}

		if err := json.NewEncoder(w).Encode(payload); err != nil {
			t.Fatalf("encode payload: %v", err)
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "sk_test", server.Client(), time.Hour)

	snapshots, err := client.FetchSnapshots(context.Background(), []string{"aapl", "AAPL"})
	if err != nil {
		t.Fatalf("fetch snapshots: %v", err)
	}

	snapshot, ok := snapshots["AAPL"]
	if !ok {
		t.Fatalf("expected AAPL snapshot to exist")
	}

	if !snapshot.Available {
		t.Fatalf("expected AAPL snapshot to be available")
	}
	if snapshot.Coverage != 3 {
		t.Fatalf("expected coverage 3, got %d", snapshot.Coverage)
	}
	if snapshot.SourceAlignment != "mixed" {
		t.Fatalf("expected mixed alignment, got %q", snapshot.SourceAlignment)
	}
	if snapshot.AverageBuzz <= 31.6 || snapshot.AverageBuzz >= 31.7 {
		t.Fatalf("expected average buzz around 31.67, got %.2f", snapshot.AverageBuzz)
	}
	if snapshot.BullishPercent <= 63.3 || snapshot.BullishPercent >= 63.4 {
		t.Fatalf("expected bullish percent around 63.33, got %.2f", snapshot.BullishPercent)
	}

	_, err = client.FetchSnapshots(context.Background(), []string{"AAPL"})
	if err != nil {
		t.Fatalf("fetch snapshots from cache: %v", err)
	}

	for _, path := range []string{
		"/reddit/stocks/v1/compare",
		"/x/stocks/v1/compare",
		"/news/stocks/v1/compare",
		"/polymarket/stocks/v1/compare",
	} {
		if requestsByPath[path] != 1 {
			t.Fatalf("expected %s to be called once, got %d", path, requestsByPath[path])
		}
	}
}

func TestFetchSnapshotsReturnsEmptyWhenDisabled(t *testing.T) {
	t.Helper()

	client := NewClient("https://api.adanos.org", "", nil, time.Minute)
	snapshots, err := client.FetchSnapshots(context.Background(), []string{"AAPL"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(snapshots) != 0 {
		t.Fatalf("expected no snapshots, got %d", len(snapshots))
	}
}

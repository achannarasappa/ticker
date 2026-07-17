package unary_test

// responseTickerBitcoinFixture mirrors the response from GET /v1/tickers/btc-bitcoin
var responseTickerBitcoinFixture = map[string]interface{}{
	"id":     "btc-bitcoin",
	"name":   "Bitcoin",
	"symbol": "BTC",
	"rank":   1,
	"quotes": map[string]interface{}{
		"USD": map[string]interface{}{
			"price":              50000.0,
			"volume_24h":         21359052876.635258,
			"market_cap":         1264515966773.0,
			"percent_change_24h": 2.5,
		},
	},
}

// responseTickerEthereumFixture mirrors the response from GET /v1/tickers/eth-ethereum
var responseTickerEthereumFixture = map[string]interface{}{
	"id":     "eth-ethereum",
	"name":   "Ethereum",
	"symbol": "ETH",
	"rank":   2,
	"quotes": map[string]interface{}{
		"USD": map[string]interface{}{
			"price":              4000.0,
			"volume_24h":         11359052876.5,
			"market_cap":         464515966773.0,
			"percent_change_24h": -1.25,
		},
	},
}

// responseSearchBitcoinFixture mirrors the response from GET /v1/search/?q=btc&c=currencies&modifier=symbol_search
var responseSearchBitcoinFixture = map[string]interface{}{
	"currencies": []map[string]interface{}{
		{
			"id":        "btcb-binance-bitcoin",
			"name":      "Binance Bitcoin",
			"symbol":    "BTCB",
			"rank":      28,
			"is_active": true,
			"type":      "token",
		},
		{
			"id":        "btc-bitcoin-inactive-copy",
			"name":      "Bitcoin Inactive Copy",
			"symbol":    "BTC",
			"rank":      0,
			"is_active": false,
			"type":      "token",
		},
		{
			"id":        "btc-bitcoin-lower-ranked-copy",
			"name":      "Bitcoin Lower Ranked Copy",
			"symbol":    "BTC",
			"rank":      5000,
			"is_active": true,
			"type":      "token",
		},
		{
			"id":        "btc-bitcoin",
			"name":      "Bitcoin",
			"symbol":    "BTC",
			"rank":      1,
			"is_active": true,
			"type":      "coin",
		},
	},
}

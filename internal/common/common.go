package common

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"
)

type Context struct {
	Config    Config
	Reference Reference
}

type Config struct {
	RefreshInterval       int      `yaml:"interval"`
	Watchlist             []string `yaml:"watchlist"`
	Lots                  []Lot    `yaml:"lots"`
	Separate              bool     `yaml:"show-separator"`
	ExtraInfoExchange     bool     `yaml:"show-tags"`
	ExtraInfoFundamentals bool     `yaml:"show-fundamentals"`
	ShowSummary           bool     `yaml:"show-summary"`
	ShowHoldings          bool     `yaml:"show-holdings"`
	Proxy                 string   `yaml:"proxy"`
	Sort                  string   `yaml:"sort"`
	Currency              string   `yaml:"currency"`
}

type Reference struct {
	CurrencyRates CurrencyRates
}

type Dependencies struct {
	Fs         afero.Fs
	HttpClient *resty.Client
}

type Lot struct {
	Symbol   string  `yaml:"symbol"`
	UnitCost float64 `yaml:"unit_cost"`
	Quantity float64 `yaml:"quantity"`
}

type CurrencyRates map[string]CurrencyRate

type CurrencyRate struct {
	FromCurrency string
	ToCurrency   string
	Rate         float64
}

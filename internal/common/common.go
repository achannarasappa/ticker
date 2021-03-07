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
	RefreshInterval       int               `yaml:"interval"`
	Watchlist             []string          `yaml:"watchlist"`
	Lots                  []Lot             `yaml:"lots"`
	Separate              bool              `yaml:"show-separator"`
	ExtraInfoExchange     bool              `yaml:"show-tags"`
	ExtraInfoFundamentals bool              `yaml:"show-fundamentals"`
	ShowSummary           bool              `yaml:"show-summary"`
	ShowHoldings          bool              `yaml:"show-holdings"`
	Proxy                 string            `yaml:"proxy"`
	Sort                  string            `yaml:"sort"`
	Currency              string            `yaml:"currency"`
	ColorScheme           ConfigColorScheme `yaml:"colors"`
}

type ConfigColorScheme struct {
	Text          string `yaml:"text"`
	TextLight     string `yaml:"text-light"`
	TextLabel     string `yaml:"text-label"`
	TextLine      string `yaml:"text-line"`
	TextTag       string `yaml:"text-tag"`
	BackgroundTag string `yaml:"background-tag"`
}

type Reference struct {
	CurrencyRates CurrencyRates
	Styles        Styles
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

type Styles struct {
	Text      StyleFn
	TextLight StyleFn
	TextLabel StyleFn
	TextBold  StyleFn
	TextLine  StyleFn
	TextPrice func(float64, string) string
	Tag       StyleFn
}

type StyleFn func(string) string

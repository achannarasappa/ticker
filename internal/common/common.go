package common

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"
)

// Context represents user defined configuration and derived reference configuration
type Context struct {
	Config    Config
	Reference Reference
}

// Config represents user defined configuration
type Config struct {
	RefreshInterval                   int               `yaml:"interval"`
	Watchlist                         []string          `yaml:"watchlist"`
	Lots                              []Lot             `yaml:"lots"`
	Separate                          bool              `yaml:"show-separator"`
	ExtraInfoExchange                 bool              `yaml:"show-tags"`
	ExtraInfoFundamentals             bool              `yaml:"show-fundamentals"`
	ShowSummary                       bool              `yaml:"show-summary"`
	ShowHoldings                      bool              `yaml:"show-holdings"`
	Proxy                             string            `yaml:"proxy"`
	Sort                              string            `yaml:"sort"`
	Currency                          string            `yaml:"currency"`
	CurrencyConvertSummaryOnly        bool              `yaml:"currency-summary-only"`
	CurrencyDisableUnitCostConversion bool              `yaml:"currency-disable-unit-cost-conversion"`
	ColorScheme                       ConfigColorScheme `yaml:"colors"`
}

// ConfigColorScheme represents user defined color scheme
type ConfigColorScheme struct {
	Text          string `yaml:"text"`
	TextLight     string `yaml:"text-light"`
	TextLabel     string `yaml:"text-label"`
	TextLine      string `yaml:"text-line"`
	TextTag       string `yaml:"text-tag"`
	BackgroundTag string `yaml:"background-tag"`
}

// Reference represents derived configuration for internal use from user defined configuration
type Reference struct {
	CurrencyRates CurrencyRates
	Styles        Styles
}

// Dependencies represents references to external dependencies
type Dependencies struct {
	Fs         afero.Fs
	HttpClient *resty.Client
}

// Lot represents a cost basis lot
type Lot struct {
	Symbol    string  `yaml:"symbol"`
	UnitCost  float64 `yaml:"unit_cost"`
	Quantity  float64 `yaml:"quantity"`
	FixedCost float64 `yaml:"fixed_cost"`
}

// CurrencyRates is a map of currency rates for lookup by currency that needs to be converted
type CurrencyRates map[string]CurrencyRate

// CurrencyRate represents a single currency conversion pair
type CurrencyRate struct {
	FromCurrency string
	ToCurrency   string
	Rate         float64
}

// Styles represents style functions for components of the UI
type Styles struct {
	Text      StyleFn
	TextLight StyleFn
	TextLabel StyleFn
	TextBold  StyleFn
	TextLine  StyleFn
	TextPrice func(float64, string) string
	Tag       StyleFn
}

// StyleFn is a function that styles text
type StyleFn func(string) string

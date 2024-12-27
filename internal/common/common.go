package common

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/afero"
)

// Context represents user defined configuration and derived reference configuration
type Context struct {
	Config    Config
	Groups    []AssetGroup
	Reference Reference
}

// Config represents user defined configuration
type Config struct {
	RefreshInterval                   int                `yaml:"interval"`
	Watchlist                         []string           `yaml:"watchlist"`
	Lots                              []Lot              `yaml:"lots"`
	Separate                          bool               `yaml:"show-separator"`
	ExtraInfoExchange                 bool               `yaml:"show-tags"`
	ExtraInfoFundamentals             bool               `yaml:"show-fundamentals"`
	ShowSummary                       bool               `yaml:"show-summary"`
	ShowHoldings                      bool               `yaml:"show-holdings"`
	Proxy                             string             `yaml:"proxy"`
	Sort                              string             `yaml:"sort"`
	Currency                          string             `yaml:"currency"`
	CurrencyConvertSummaryOnly        bool               `yaml:"currency-summary-only"`
	CurrencyDisableUnitCostConversion bool               `yaml:"currency-disable-unit-cost-conversion"`
	ColorScheme                       ConfigColorScheme  `yaml:"colors"`
	AssetGroup                        []ConfigAssetGroup `yaml:"groups"`
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

type ConfigAssetGroup struct {
	Name      string   `yaml:"name"`
	Watchlist []string `yaml:"watchlist"`
	Holdings  []Lot    `yaml:"holdings"`
}

type AssetGroup struct {
	ConfigAssetGroup
	SymbolsBySource []AssetGroupSymbolsBySource
}

type AssetGroupSymbolsBySource struct {
	Symbols []string
	Source  QuoteSource
}

type AssetGroupQuote struct {
	AssetGroup  AssetGroup
	AssetQuotes []AssetQuote
}

// Reference represents derived configuration for internal use from user defined configuration
type Reference struct {
	CurrencyRates                  CurrencyRates
	SourceToUnderlyingAssetSymbols map[QuoteSource][]string
	Styles                         Styles
}

// Dependencies represents references to external dependencies
type Dependencies struct {
	Fs          afero.Fs
	HttpClients DependenciesHttpClients //nolint:golint,stylecheck,revive
}

type DependenciesHttpClients struct { //nolint:golint,stylecheck,revive
	Default      *resty.Client
	Yahoo        *resty.Client
	YahooSession *resty.Client
}

// Lot represents a cost basis lot
type Lot struct {
	Symbol    string  `yaml:"symbol"`
	UnitCost  float64 `yaml:"unit_cost"`
	Quantity  float64 `yaml:"quantity"`
	FixedCost float64 `yaml:"fixed_cost"`
	// FixedProperties LotFixedProperties `yaml:"fixed_properties"`
}

// type LotFixedProperties struct {
// 	Class       string  `yaml:"class"`
// 	Description string  `yaml:"description"`
// 	Currency    string  `yaml:"currency"`
// 	UnitValue   float64 `yaml:"unit_value"`
// }

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

type HoldingChange struct {
	Amount  float64
	Percent float64
}

type Meta struct {
	IsVariablePrecision bool
	OrderIndex          int
}

type Holding struct {
	Value       float64
	Cost        float64
	Quantity    float64
	UnitValue   float64
	UnitCost    float64
	DayChange   HoldingChange
	TotalChange HoldingChange
	Weight      float64
}

// Currency is the original and converted currency if applicable
type Currency struct {
	// Code is the original currency code of the asset
	FromCurrencyCode string
	// CodeConverted is the currency code that pricing and values have been converted into
	ToCurrencyCode string
}

type QuotePrice struct {
	Price          float64
	PricePrevClose float64
	PriceOpen      float64
	PriceDayHigh   float64
	PriceDayLow    float64
	Change         float64
	ChangePercent  float64
}

type QuoteExtended struct {
	FiftyTwoWeekHigh float64
	FiftyTwoWeekLow  float64
	MarketCap        float64
	Volume           float64
}

type QuoteFutures struct {
	SymbolUnderlying string
	IndexPrice       float64
	Basis            float64
	OpenInterest     float64
	Expiry           string
}

type Exchange struct {
	Name                    string
	Delay                   float64
	State                   ExchangeState
	IsActive                bool
	IsRegularTradingSession bool
}

type ExchangeState int

const (
	ExchangeStateOpen ExchangeState = iota
	ExchangeStatePremarket
	ExchangeStatePostmarket
	ExchangeStateClosed
)

type Asset struct {
	Name          string
	Symbol        string
	Class         AssetClass
	Currency      Currency
	Holding       Holding
	QuotePrice    QuotePrice
	QuoteExtended QuoteExtended
	QuoteFutures  QuoteFutures
	QuoteSource   QuoteSource
	Exchange      Exchange
	Meta          Meta
}

type AssetClass int

const (
	AssetClassCash AssetClass = iota
	AssetClassStock
	AssetClassCryptocurrency
	AssetClassPrivateSecurity
	AssetClassUnknown
	AssetClassFuturesContract
)

type QuoteSource int

const (
	QuoteSourceYahoo QuoteSource = iota
	QuoteSourceUserDefined
	QuoteSourceCoingecko
	QuoteSourceUnknown
	QuoteSourceCoinCap
	QuoteSourceCoinbase
)

// AssetQuote represents a price quote and related attributes for a single security
type AssetQuote struct {
	Name          string
	Symbol        string
	Class         AssetClass
	Currency      Currency
	QuotePrice    QuotePrice
	QuoteExtended QuoteExtended
	QuoteFutures  QuoteFutures
	QuoteSource   QuoteSource
	Exchange      Exchange
	Meta          Meta
}

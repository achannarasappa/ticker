package common

import (
	"log"
	"strings"
	"time"

	"github.com/spf13/afero"
)

// Context represents user defined configuration and derived reference configuration
type Context struct {
	Config    Config
	Groups    []AssetGroup
	Reference Reference
	Logger    *log.Logger
	Cache     Cache
}

// Cache is a key/value store for data fetched at startup and other
// slow-changing data, shared across ticker instances on the same machine.
// Implementations must tolerate being nil-checked by callers and must never
// panic on a miss.
type Cache interface {
	// Get reports whether key has a fresh entry and, if so, decodes its value
	// into out. It returns false on a miss, an expired entry, or a decode error.
	Get(key string, out any) bool
	// Set stores value under key with the given time-to-live. Failures are
	// silently ignored since the cache is an optimization that must never break
	// startup.
	Set(key string, value any, ttl time.Duration)
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
	ShowHoldings                      bool               `yaml:"show-holdings"`  // Deprecated: use ShowPositions instead, kept for backwards compatibility
	ShowPositions                     bool               `yaml:"show-positions"` // Preferred field name
	Sort                              string             `yaml:"sort"`
	Currency                          string             `yaml:"currency"`
	CurrencyConvertSummaryOnly        bool               `yaml:"currency-summary-only"`
	CurrencyDisableUnitCostConversion bool               `yaml:"currency-disable-unit-cost-conversion"`
	ColorScheme                       ConfigColorScheme  `yaml:"colors"`
	AssetGroup                        []ConfigAssetGroup `yaml:"groups"`
	SnapTrade                         ConfigSnapTrade    `yaml:"snaptrade"`
	Debug                             bool               `yaml:"debug"`
	// Cache enables the on-disk cache. It is a pointer so that an unset config
	// value (nil) can be distinguished from an explicit false, allowing the
	// cache to default to on while still being disableable via config or
	// --no-cache.
	Cache *bool `yaml:"cache"`
}

// ConfigSnapTrade represents SnapTrade credentials used to pull live brokerage holdings.
// Personal keys (a single self-serve user) need only ClientID + ConsumerKey. Commercial
// keys additionally register and identify users via UserID.
type ConfigSnapTrade struct {
	ClientID    string `yaml:"client-id"`
	ConsumerKey string `yaml:"consumer-key"` // secret
	AccountType string `yaml:"account-type"` // "personal" or "commercial"; inferred from UserID when empty
	UserID      string `yaml:"user-id"`      // commercial only: partner-chosen user identifier, e.g. an email
}

// IsPersonal reports whether these are personal SnapTrade keys (the client ID and
// consumer key identify a single user, so no registration or user ID is needed).
// An explicit AccountType wins; otherwise the absence of a UserID implies personal.
func (st ConfigSnapTrade) IsPersonal() bool {
	if st.AccountType != "" {
		return strings.EqualFold(st.AccountType, "personal")
	}

	return st.UserID == ""
}

// ConfigColorScheme represents user defined color scheme
type ConfigColorScheme struct {
	Text             string `yaml:"text"`
	TextLight        string `yaml:"text-light"`
	TextLabel        string `yaml:"text-label"`
	TextLine         string `yaml:"text-line"`
	TextTag          string `yaml:"text-tag"`
	TextHeader       string `yaml:"text-header"`
	BackgroundTag    string `yaml:"background-tag"`
	BackgroundHeader string `yaml:"background-header"`
}

type ConfigAssetGroup struct {
	Name      string   `yaml:"name"`
	Watchlist []string `yaml:"watchlist"`
	Lots      []Lot    `yaml:"lots"`     // Preferred field name
	Holdings  []Lot    `yaml:"holdings"` // Deprecated: use Lots instead, kept for backwards compatibility
	Options   []Option `yaml:"options"`
}

type AssetGroup struct {
	ConfigAssetGroup
	SymbolsBySource []AssetGroupSymbolsBySource
	// IsSnapTrade marks groups derived from a SnapTrade brokerage account (built at runtime, not from config)
	IsSnapTrade bool
	// SnapTradeAccountID identifies the account so its holdings can be loaded lazily
	SnapTradeAccountID string
	// SnapTradeInstitution is the brokerage name (e.g. "Robinhood"), shown as the group source
	SnapTradeInstitution string
	// SnapTradeAccountNumber is the account number, shown masked (last 4) alongside the name
	SnapTradeAccountNumber string
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
	Styles Styles
}

// Dependencies represents references to external dependencies
type Dependencies struct {
	Fs                               afero.Fs
	SymbolsURL                       string
	GitHubReleasesURL                string
	MonitorPriceCoinbaseBaseURL      string
	MonitorPriceCoinbaseStreamingURL string
	MonitorYahooBaseURL              string
	MonitorYahooSessionRootURL       string
	MonitorYahooSessionCrumbURL      string
	MonitorYahooSessionConsentURL    string
	SnapTradeBaseURL                 string
}

type Monitor interface {
	Start() error
	GetAssetQuotes(ignoreCache ...bool) ([]AssetQuote, error)
	SetSymbols(symbols []string, versionVector int) error
	SetCurrencyRates(currencyRates CurrencyRates) error
	Stop() error
}

type MonitorCurrencyRate interface {
	Start() error
	SetTargetCurrency(targetCurrency string)
	Stop() error
}

// Lot represents a cost basis lot
type Lot struct {
	Symbol    string  `yaml:"symbol"`
	UnitCost  float64 `yaml:"unit_cost"`
	Quantity  float64 `yaml:"quantity"`
	FixedCost float64 `yaml:"fixed_cost"`
}

// Option represents an option contract
type Option struct {
	Symbol         string  `yaml:"symbol"`
	StrikePrice    float64 `yaml:"strike_price"`
	Type           string  `yaml:"type"` // "put" or "call"
	Premium        float64 `yaml:"premium"`
	CurrentPremium float64 `yaml:"current_premium"` // current market premium per share (optional; from broker)
	Contracts      float64 `yaml:"contracts"`
	Expiration     string  `yaml:"expiration"`
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
	Text       StyleFn
	TextLight  StyleFn
	TextLabel  StyleFn
	TextBold   StyleFn
	TextLine   StyleFn
	TextHeader StyleFn
	TextPrice  func(float64, string) string
	Tag        StyleFn
}

// StyleFn is a function that styles text
type StyleFn func(string) string

type PositionChange struct {
	Amount  float64
	Percent float64
}

type Meta struct {
	IsVariablePrecision bool
	OrderIndex          int
	SymbolInSourceAPI   string
}

type Position struct {
	Value       float64
	Cost        float64
	Quantity    float64
	UnitValue   float64
	UnitCost    float64
	DayChange   PositionChange
	TotalChange PositionChange
	Weight      float64
}

// Currency is the original and converted currency if applicable
type Currency struct {
	// Code is the original currency code of the asset
	FromCurrencyCode string
	// CodeConverted is the currency code that pricing and values have been converted into
	ToCurrencyCode string
	// Rate is the conversion rate from the original currency to the converted currency
	Rate float64
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
	ContractSize     float64
}

type QuoteOption struct {
	StrikePrice    float64
	BreakevenPrice float64
	Type           string // "put" or "call"
	Premium        float64
	CurrentPremium float64
	Contracts      float64
	DiffToStrike   float64
	Expiration     string
}

type Exchange struct {
	Name                    string
	Delay                   float64
	DelayText               string
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
	Position      Position
	QuotePrice    QuotePrice
	QuoteExtended QuoteExtended
	QuoteFutures  QuoteFutures
	QuoteOption   QuoteOption
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
	AssetClassOption
	AssetClassCurrency
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

type MessageUpdate[T any] struct {
	Data          T
	ID            string
	Sequence      int64
	VersionVector int
}

type MessageRequest[T any] struct {
	Data          T
	ID            string
	VersionVector int
}

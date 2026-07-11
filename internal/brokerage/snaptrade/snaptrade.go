// Package snaptrade provides a thin client for the SnapTrade API used to pull
// live brokerage holdings (e.g. Robinhood) into ticker as asset groups.
package snaptrade

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	c "github.com/achannarasappa/ticker/v5/internal/common"
)

// optionContractMultiplier is the standard number of shares per US equity option
// contract. SnapTrade reports option average_purchase_price per contract, so we
// divide by this to get the per-share premium ticker's breakeven math expects.
const optionContractMultiplier = 100

// Client is a minimal SnapTrade API client. It mirrors the request signing
// performed by the official SnapTrade SDKs: an HMAC-SHA256 over a canonical
// JSON object of {content, path, query}, base64-encoded into a Signature header.
type Client struct {
	client      *http.Client
	baseURL     string
	clientID    string
	consumerKey string
}

// New returns a SnapTrade client. baseURL is injected for testability
// (production uses https://api.snaptrade.com/api/v1).
func New(baseURL, clientID, consumerKey string) *Client {
	return &Client{
		client:      &http.Client{Timeout: 10 * time.Second},
		baseURL:     baseURL,
		clientID:    clientID,
		consumerKey: consumerKey,
	}
}

// Account is a brokerage account returned by SnapTrade.
type Account struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Number          string `json:"number"`
	InstitutionName string `json:"institution_name"`
}

// Position is a single holding within an account.
type Position struct {
	Symbol               *PositionSymbol `json:"symbol"`
	Units                float64         `json:"units"`
	AveragePurchasePrice float64         `json:"average_purchase_price"`
}

// PositionSymbol and UniversalSymbol model SnapTrade's nested symbol object
// (position.symbol.symbol.symbol is the ticker).
type PositionSymbol struct {
	Symbol *UniversalSymbol `json:"symbol"`
}

type UniversalSymbol struct {
	Symbol string `json:"symbol"`
}

// OptionsPosition is a single option holding within an account.
type OptionsPosition struct {
	Symbol               *OptionBrokerageSymbol `json:"symbol"`
	Units                float64                `json:"units"`
	Price                float64                `json:"price"` // current market price per contract
	AveragePurchasePrice float64                `json:"average_purchase_price"`
}

// OptionBrokerageSymbol, OptionsSymbol, and UnderlyingSymbol model SnapTrade's
// nested option symbol (symbol.option_symbol describes the contract; its
// underlying_symbol.symbol is the ticker used for live pricing).
type OptionBrokerageSymbol struct {
	OptionSymbol *OptionsSymbol `json:"option_symbol"`
}

type OptionsSymbol struct {
	OptionType       string            `json:"option_type"` // "CALL" or "PUT"
	StrikePrice      float64           `json:"strike_price"`
	ExpirationDate   string            `json:"expiration_date"`
	UnderlyingSymbol *UnderlyingSymbol `json:"underlying_symbol"`
}

type UnderlyingSymbol struct {
	Symbol string `json:"symbol"`
}

type registerResponse struct {
	UserID     string `json:"userId"`
	UserSecret string `json:"userSecret"`
}

type loginResponse struct {
	RedirectURI string `json:"redirectURI"`
}

// RegisterUser registers a partner user with SnapTrade and returns the userSecret
// that must be persisted and used for all subsequent requests for that user.
func (c *Client) RegisterUser(userID string) (string, error) {
	body := map[string]string{"userId": userID}

	respBytes, err := c.doRequest(http.MethodPost, "/snapTrade/registerUser", nil, body)
	if err != nil {
		return "", err
	}

	var out registerResponse
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return "", fmt.Errorf("snaptrade: decode registerUser: %w", err)
	}

	return out.UserSecret, nil
}

// GetLoginRedirectURI returns the connection-portal URL a user opens in a browser
// to authorize a brokerage (e.g. Robinhood).
func (c *Client) GetLoginRedirectURI(userID, userSecret string) (string, error) {
	respBytes, err := c.doRequest(http.MethodPost, "/snapTrade/login", userQuery(userID, userSecret), nil)
	if err != nil {
		return "", err
	}

	var out loginResponse
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return "", fmt.Errorf("snaptrade: decode login: %w", err)
	}

	return out.RedirectURI, nil
}

// ListAccounts returns every brokerage account connected by the user.
func (c *Client) ListAccounts(userID, userSecret string) ([]Account, error) {
	respBytes, err := c.doRequest(http.MethodGet, "/accounts", userQuery(userID, userSecret), nil)
	if err != nil {
		return nil, err
	}

	var out []Account
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return nil, fmt.Errorf("snaptrade: decode accounts: %w", err)
	}

	return out, nil
}

// GetPositions returns the equity positions held in a single account.
func (c *Client) GetPositions(userID, userSecret, accountID string) ([]Position, error) {
	subpath := "/accounts/" + url.PathEscape(accountID) + "/positions"

	respBytes, err := c.doRequest(http.MethodGet, subpath, userQuery(userID, userSecret), nil)
	if err != nil {
		return nil, err
	}

	var out []Position
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return nil, fmt.Errorf("snaptrade: decode positions: %w", err)
	}

	return out, nil
}

// ListOptionHoldings returns the option positions held in a single account.
func (c *Client) ListOptionHoldings(userID, userSecret, accountID string) ([]OptionsPosition, error) {
	subpath := "/accounts/" + url.PathEscape(accountID) + "/options"

	respBytes, err := c.doRequest(http.MethodGet, subpath, userQuery(userID, userSecret), nil)
	if err != nil {
		return nil, err
	}

	var out []OptionsPosition
	if err := json.Unmarshal(respBytes, &out); err != nil {
		return nil, fmt.Errorf("snaptrade: decode option holdings: %w", err)
	}

	return out, nil
}

// userQuery builds the user-scoped query params. For personal keys both values
// are empty and omitted — the client ID and consumer key identify the user.
func userQuery(userID, userSecret string) url.Values {
	query := url.Values{}

	if userID != "" {
		query.Set("userId", userID)
	}

	if userSecret != "" {
		query.Set("userSecret", userSecret)
	}

	return query
}

// doRequest builds, signs, and sends a request, returning the response body.
// clientId and timestamp are added to every request's query string.
func (c *Client) doRequest(method, subpath string, query url.Values, body interface{}) ([]byte, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("clientId", c.clientID)
	query.Set("timestamp", strconv.FormatInt(time.Now().Unix(), 10))

	var bodyBytes []byte
	if body != nil {
		var err error
		if bodyBytes, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}

	requestURL, err := url.Parse(c.baseURL + subpath)
	if err != nil {
		return nil, err
	}
	requestURL.RawQuery = query.Encode()

	signature, err := c.sign(requestURL.Path, requestURL.RawQuery, bodyBytes)
	if err != nil {
		return nil, err
	}

	var reqBody io.Reader
	if bodyBytes != nil {
		reqBody = bytes.NewReader(bodyBytes)
	}

	req, err := http.NewRequest(method, requestURL.String(), reqBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Signature", signature)
	req.Header.Set("Accept", "application/json")
	if bodyBytes != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("snaptrade: unexpected status %d: %s", resp.StatusCode, string(respBytes))
	}

	return respBytes, nil
}

// sign computes the SnapTrade request signature: HMAC-SHA256 (keyed by the
// consumer key) over the canonical JSON encoding of {content, path, query},
// base64-encoded. content is the parsed request body, or null when there is none.
func (c *Client) sign(path, rawQuery string, body []byte) (string, error) {
	var content interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &content); err != nil {
			return "", err
		}
	}

	// Encoding a map sorts keys alphabetically (content, path, query) and, with
	// HTML escaping disabled, matches the canonical form the SnapTrade API expects.
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(map[string]interface{}{
		"content": content,
		"path":    path,
		"query":   rawQuery,
	}); err != nil {
		return "", err
	}

	mac := hmac.New(sha256.New, []byte(c.consumerKey))
	mac.Write([]byte(strings.TrimSuffix(buf.String(), "\n")))

	return base64.StdEncoding.EncodeToString(mac.Sum(nil)), nil
}

// TransformToConfigAssetGroup maps a SnapTrade account, its equity positions,
// and its option positions into a single ticker asset group. Holdings and
// options are keyed by ticker/underlying symbol; prices are resolved live by
// the normal quote sources, and only quantity and cost basis come from SnapTrade.
func TransformToConfigAssetGroup(account Account, positions []Position, optionPositions []OptionsPosition) c.ConfigAssetGroup {
	holdings := make([]c.Lot, 0, len(positions))

	for _, position := range positions {
		if lot, ok := transformPositionToLot(position); ok {
			holdings = append(holdings, lot)
		}
	}

	options := make([]c.Option, 0, len(optionPositions))

	for _, position := range optionPositions {
		if option, ok := transformOptionPosition(position); ok {
			options = append(options, option)
		}
	}

	return c.ConfigAssetGroup{
		Name:     AccountName(account),
		Holdings: holdings,
		Options:  options,
	}
}

func transformOptionPosition(position OptionsPosition) (c.Option, bool) {
	if position.Symbol == nil || position.Symbol.OptionSymbol == nil {
		return c.Option{}, false
	}

	optionSymbol := position.Symbol.OptionSymbol
	if optionSymbol.UnderlyingSymbol == nil || optionSymbol.UnderlyingSymbol.Symbol == "" {
		return c.Option{}, false
	}

	return c.Option{
		Symbol:      optionSymbol.UnderlyingSymbol.Symbol,
		StrikePrice: optionSymbol.StrikePrice,
		Type:        strings.ToLower(optionSymbol.OptionType),
		// SnapTrade reports average_purchase_price per contract but the current
		// price per share, so only the former is divided by the contract multiplier.
		Premium:        position.AveragePurchasePrice / optionContractMultiplier,
		CurrentPremium: position.Price,
		Contracts:      position.Units,
		Expiration:     optionSymbol.ExpirationDate,
	}, true
}

func transformPositionToLot(position Position) (c.Lot, bool) {
	ticker := positionTicker(position)
	if ticker == "" {
		return c.Lot{}, false
	}

	return c.Lot{
		Symbol:   ticker,
		Quantity: position.Units,
		UnitCost: position.AveragePurchasePrice,
	}, true
}

func positionTicker(position Position) string {
	if position.Symbol == nil || position.Symbol.Symbol == nil {
		return ""
	}

	return position.Symbol.Symbol.Symbol
}

// AccountName returns a display name for an account, falling back to its number then id.
func AccountName(account Account) string {
	if account.Name != "" {
		return account.Name
	}

	if account.Number != "" {
		return account.Number
	}

	return account.ID
}

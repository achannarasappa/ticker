package snaptrade

import (
	"net/http"

	c "github.com/achannarasappa/ticker/v5/internal/common"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("SnapTrade Client", func() {

	var server *ghttp.Server

	BeforeEach(func() {
		server = ghttp.NewServer()
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("RegisterUser", func() {
		It("should register the user and return the user secret", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/snapTrade/registerUser"),
					ghttp.VerifyHeader(http.Header{"Content-Type": []string{"application/json"}}),
					verifySignaturePresent(),
					ghttp.RespondWithJSONEncoded(http.StatusOK, registerResponse{UserID: "ben", UserSecret: "user-secret-123"}),
				),
			)

			secret, err := New(server.URL(), "client-id", "consumer-key").RegisterUser("ben")

			Expect(err).NotTo(HaveOccurred())
			Expect(secret).To(Equal("user-secret-123"))
		})
	})

	Describe("GetLoginRedirectURI", func() {
		It("should return the connection portal URL", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/snapTrade/login"),
					verifyQueryParam("userId", "ben"),
					verifyQueryParam("userSecret", "secret"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, loginResponse{RedirectURI: "https://app.snaptrade.com/connect/abc"}),
				),
			)

			uri, err := New(server.URL(), "client-id", "consumer-key").GetLoginRedirectURI("ben", "secret")

			Expect(err).NotTo(HaveOccurred())
			Expect(uri).To(Equal("https://app.snaptrade.com/connect/abc"))
		})
	})

	Describe("ListAccounts", func() {
		It("should return the connected accounts", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/accounts"),
					verifySignaturePresent(),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []Account{
						{ID: "acct-1", Name: "Robinhood Individual", Number: "RH123"},
					}),
				),
			)

			accounts, err := New(server.URL(), "client-id", "consumer-key").ListAccounts("ben", "secret")

			Expect(err).NotTo(HaveOccurred())
			Expect(accounts).To(HaveLen(1))
			Expect(accounts[0].Name).To(Equal("Robinhood Individual"))
		})

		When("using personal keys (empty user id and secret)", func() {
			It("should omit the userId and userSecret query params", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/accounts"),
						func(_ http.ResponseWriter, req *http.Request) {
							defer GinkgoRecover()
							Expect(req.URL.Query().Has("userId")).To(BeFalse())
							Expect(req.URL.Query().Has("userSecret")).To(BeFalse())
							Expect(req.URL.Query().Get("clientId")).To(Equal("client-id"))
						},
						ghttp.RespondWithJSONEncoded(http.StatusOK, []Account{{ID: "acct-1", Name: "Robinhood"}}),
					),
				)

				accounts, err := New(server.URL(), "client-id", "consumer-key").ListAccounts("", "")

				Expect(err).NotTo(HaveOccurred())
				Expect(accounts).To(HaveLen(1))
			})
		})
	})

	Describe("GetPositions", func() {
		It("should return the positions for an account", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/accounts/acct-1/positions"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []Position{
						{
							Symbol:               &PositionSymbol{Symbol: &UniversalSymbol{Symbol: "AAPL"}},
							Units:                10,
							AveragePurchasePrice: 150,
						},
					}),
				),
			)

			positions, err := New(server.URL(), "client-id", "consumer-key").GetPositions("ben", "secret", "acct-1")

			Expect(err).NotTo(HaveOccurred())
			Expect(positions).To(HaveLen(1))
			Expect(positions[0].Units).To(Equal(10.0))
		})

		When("the server returns a non-2xx status", func() {
			It("should return an error", func() {
				server.AppendHandlers(
					ghttp.RespondWith(http.StatusUnauthorized, `{"detail":"bad signature"}`),
				)

				_, err := New(server.URL(), "client-id", "consumer-key").GetPositions("ben", "secret", "acct-1")

				Expect(err).To(HaveOccurred())
			})
		})
	})

	Describe("ListOptionHoldings", func() {
		It("should return the option positions for an account", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/accounts/acct-1/options"),
					ghttp.RespondWithJSONEncoded(http.StatusOK, []OptionsPosition{
						{
							Symbol: &OptionBrokerageSymbol{OptionSymbol: &OptionsSymbol{
								OptionType:       "CALL",
								StrikePrice:      455,
								UnderlyingSymbol: &UnderlyingSymbol{Symbol: "TSLA"},
							}},
							Units:                2,
							AveragePurchasePrice: 6.5,
						},
					}),
				),
			)

			positions, err := New(server.URL(), "client-id", "consumer-key").ListOptionHoldings("ben", "secret", "acct-1")

			Expect(err).NotTo(HaveOccurred())
			Expect(positions).To(HaveLen(1))
			Expect(positions[0].Symbol.OptionSymbol.StrikePrice).To(Equal(455.0))
		})
	})

	Describe("TransformToConfigAssetGroup with options", func() {
		It("should map both equity and option positions into a single group", func() {
			positions := []Position{
				{Symbol: &PositionSymbol{Symbol: &UniversalSymbol{Symbol: "NVDA"}}, Units: 3, AveragePurchasePrice: 4.42},
			}
			optionPositions := []OptionsPosition{
				{
					Symbol: &OptionBrokerageSymbol{OptionSymbol: &OptionsSymbol{
						OptionType:       "PUT",
						StrikePrice:      255,
						ExpirationDate:   "2026-07-10",
						UnderlyingSymbol: &UnderlyingSymbol{Symbol: "AAPL"},
					}},
					Units:                1,
					Price:                4.2,
					AveragePurchasePrice: 270,
				},
			}

			group := TransformToConfigAssetGroup(Account{ID: "acct-1", Name: "Robinhood"}, positions, optionPositions)

			Expect(group.Name).To(Equal("Robinhood"))
			Expect(group.Holdings).To(Equal([]c.Lot{{Symbol: "NVDA", Quantity: 3, UnitCost: 4.42}}))
			// average_purchase_price is per contract (÷100); price is already per share
			Expect(group.Options).To(Equal([]c.Option{
				{Symbol: "AAPL", StrikePrice: 255, Type: "put", Premium: 2.7, CurrentPremium: 4.2, Contracts: 1, Expiration: "2026-07-10"},
			}))
		})

		It("should skip option positions without a resolvable underlying symbol", func() {
			group := TransformToConfigAssetGroup(Account{ID: "acct-1"}, nil, []OptionsPosition{{Symbol: nil}})

			Expect(group.Options).To(BeEmpty())
		})
	})

	Describe("sign", func() {
		var client *Client

		BeforeEach(func() {
			client = New("https://api.snaptrade.com/api/v1", "abc", "secret")
		})

		It("should produce the canonical HMAC signature for a request with no body", func() {
			signature, err := client.sign("/api/v1/accounts", "clientId=abc&timestamp=1", nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(signature).To(Equal("FZk8SoFXn+2QWJAy5s+KRa5Eydvu0zLrODEliNrDWYw="))
		})

		It("should include the parsed body as content", func() {
			signature, err := client.sign("/api/v1/snapTrade/registerUser", "clientId=abc&timestamp=1", []byte(`{"userId":"ben"}`))

			Expect(err).NotTo(HaveOccurred())
			Expect(signature).To(Equal("OkuZtvdJVBHxoqA8rtDhb7HkuZsnr/QqRITwU03/j2E="))
		})

		It("should differ when the consumer key differs", func() {
			a, _ := client.sign("/api/v1/accounts", "clientId=abc&timestamp=1", nil)
			b, _ := New("", "abc", "other-key").sign("/api/v1/accounts", "clientId=abc&timestamp=1", nil)

			Expect(a).NotTo(Equal(b))
		})
	})

	Describe("TransformToConfigAssetGroup", func() {
		It("should map an account and its positions to a config asset group", func() {
			account := Account{ID: "acct-1", Name: "Robinhood Individual", Number: "RH123"}
			positions := []Position{
				{Symbol: &PositionSymbol{Symbol: &UniversalSymbol{Symbol: "AAPL"}}, Units: 10, AveragePurchasePrice: 150},
				{Symbol: &PositionSymbol{Symbol: &UniversalSymbol{Symbol: "TSLA"}}, Units: 2, AveragePurchasePrice: 250},
			}

			group := TransformToConfigAssetGroup(account, positions, nil)

			Expect(group.Name).To(Equal("Robinhood Individual"))
			Expect(group.Holdings).To(Equal([]c.Lot{
				{Symbol: "AAPL", Quantity: 10, UnitCost: 150},
				{Symbol: "TSLA", Quantity: 2, UnitCost: 250},
			}))
		})

		It("should skip positions without a resolvable symbol (e.g. cash)", func() {
			positions := []Position{
				{Symbol: &PositionSymbol{Symbol: &UniversalSymbol{Symbol: "AAPL"}}, Units: 10, AveragePurchasePrice: 150},
				{Symbol: nil, Units: 100, AveragePurchasePrice: 1},
			}

			group := TransformToConfigAssetGroup(Account{ID: "acct-1"}, positions, nil)

			Expect(group.Holdings).To(HaveLen(1))
			Expect(group.Holdings[0].Symbol).To(Equal("AAPL"))
		})

		It("should fall back to account number then id for the group name", func() {
			Expect(TransformToConfigAssetGroup(Account{ID: "acct-1", Number: "RH123"}, nil, nil).Name).To(Equal("RH123"))
			Expect(TransformToConfigAssetGroup(Account{ID: "acct-1"}, nil, nil).Name).To(Equal("acct-1"))
		})
	})
})

func verifySignaturePresent() http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		defer GinkgoRecover()
		Expect(req.Header.Get("Signature")).NotTo(BeEmpty())
	}
}

func verifyQueryParam(key, value string) http.HandlerFunc {
	return func(_ http.ResponseWriter, req *http.Request) {
		defer GinkgoRecover()
		Expect(req.URL.Query().Get(key)).To(Equal(value))
	}
}

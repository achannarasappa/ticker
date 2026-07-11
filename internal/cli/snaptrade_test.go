package cli_test

import (
	"net/http"

	"github.com/adrg/xdg"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/spf13/afero"

	"github.com/achannarasappa/ticker/v5/internal/brokerage/snaptrade"
	"github.com/achannarasappa/ticker/v5/internal/cli"
	c "github.com/achannarasappa/ticker/v5/internal/common"
)

var _ = Describe("SnapTrade", func() {

	var (
		server *ghttp.Server
		fs     afero.Fs
		dep    c.Dependencies
		config c.Config
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		fs = afero.NewMemMapFs()
		dep = c.Dependencies{
			Fs:               fs,
			SymbolsURL:       server.URL() + "/symbols.csv",
			SnapTradeBaseURL: server.URL(),
		}
		config = c.Config{
			SnapTrade: c.ConfigSnapTrade{
				ClientID:    "client-id",
				ConsumerKey: "consumer-key",
				UserID:      "ben",
			},
		}

		server.RouteToHandler("GET", "/symbols.csv",
			ghttp.RespondWith(http.StatusOK, `"BTC.X","BTC-USD","cb"`, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	seedSecret := func() {
		Expect(snaptrade.NewStore(fs, xdg.DataHome).SaveUserSecret("ben", "secret")).To(Succeed())
	}

	Describe("FetchSnapTradeAccountGroups", func() {

		When("the user is connected", func() {
			It("should return a placeholder group per account (no holdings fetched yet)", func() {
				seedSecret()
				requestedPositions := false
				server.RouteToHandler("GET", "/accounts",
					ghttp.RespondWithJSONEncoded(http.StatusOK, []snaptrade.Account{
						{ID: "acct-1", Name: "Robinhood Individual"},
						{ID: "acct-2", Name: "Robinhood Crypto"},
					}),
				)
				server.RouteToHandler("GET", "/accounts/acct-1/positions", func(_ http.ResponseWriter, _ *http.Request) {
					requestedPositions = true
				})

				groups, err := cli.FetchSnapTradeAccountGroups(dep, config)

				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(HaveLen(2))
				Expect(groups[0].Name).To(Equal("Robinhood Individual"))
				Expect(groups[0].IsSnapTrade).To(BeTrue())
				Expect(groups[0].SnapTradeAccountID).To(Equal("acct-1"))
				Expect(groups[0].Holdings).To(BeEmpty())
				Expect(groups[0].SymbolsBySource).To(BeEmpty())
				Expect(requestedPositions).To(BeFalse()) // holdings are loaded lazily, not here
			})
		})

		When("using personal keys", func() {
			It("should omit the userId query param", func() {
				personalConfig := c.Config{SnapTrade: c.ConfigSnapTrade{ClientID: "client-id", ConsumerKey: "consumer-key"}}
				server.RouteToHandler("GET", "/accounts", ghttp.CombineHandlers(
					func(_ http.ResponseWriter, req *http.Request) {
						defer GinkgoRecover()
						Expect(req.URL.Query().Has("userId")).To(BeFalse())
					},
					ghttp.RespondWithJSONEncoded(http.StatusOK, []snaptrade.Account{{ID: "acct-1", Name: "Robinhood"}}),
				))

				groups, err := cli.FetchSnapTradeAccountGroups(dep, personalConfig)

				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(HaveLen(1))
			})
		})

		When("SnapTrade is not configured", func() {
			It("should return no groups", func() {
				groups, err := cli.FetchSnapTradeAccountGroups(dep, c.Config{})

				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})
		})

		When("the user has not connected (no stored secret)", func() {
			It("should return no groups", func() {
				groups, err := cli.FetchSnapTradeAccountGroups(dep, config)

				Expect(err).NotTo(HaveOccurred())
				Expect(groups).To(BeEmpty())
			})
		})
	})

	Describe("LoadSnapTradeAccountGroup", func() {

		When("the account holds stocks and options", func() {
			It("should build a single resolved group with holdings and options", func() {
				seedSecret()
				server.RouteToHandler("GET", "/accounts/acct-1/positions",
					ghttp.RespondWithJSONEncoded(http.StatusOK, []snaptrade.Position{
						{Symbol: &snaptrade.PositionSymbol{Symbol: &snaptrade.UniversalSymbol{Symbol: "NVDA"}}, Units: 10, AveragePurchasePrice: 150},
					}),
				)
				server.RouteToHandler("GET", "/accounts/acct-1/options",
					ghttp.RespondWithJSONEncoded(http.StatusOK, []snaptrade.OptionsPosition{
						{
							Symbol: &snaptrade.OptionBrokerageSymbol{OptionSymbol: &snaptrade.OptionsSymbol{
								OptionType:       "PUT",
								StrikePrice:      255,
								UnderlyingSymbol: &snaptrade.UnderlyingSymbol{Symbol: "AAPL"},
							}},
							Units:                2,
							AveragePurchasePrice: 270,
						},
					}),
				)

				group, err := cli.LoadSnapTradeAccountGroup(dep, config, "acct-1", "Robinhood Individual")

				Expect(err).NotTo(HaveOccurred())
				Expect(group.IsSnapTrade).To(BeTrue())
				Expect(group.SnapTradeAccountID).To(Equal("acct-1"))
				Expect(group.Name).To(Equal("Robinhood Individual"))
				Expect(group.Holdings).To(Equal([]c.Lot{{Symbol: "NVDA", Quantity: 10, UnitCost: 150}}))
				Expect(group.Options).To(Equal([]c.Option{
					{Symbol: "AAPL", StrikePrice: 255, Type: "put", Premium: 2.7, Contracts: 2},
				}))
				Expect(group.SymbolsBySource).NotTo(BeEmpty())
			})
		})

		When("positions cannot be fetched", func() {
			It("should return an empty group and an error", func() {
				seedSecret()
				server.RouteToHandler("GET", "/accounts/acct-1/positions",
					ghttp.RespondWith(http.StatusInternalServerError, ""),
				)

				group, err := cli.LoadSnapTradeAccountGroup(dep, config, "acct-1", "Robinhood Individual")

				Expect(err).To(HaveOccurred())
				Expect(group.SnapTradeAccountID).To(BeEmpty())
			})
		})
	})

	Describe("SnapTrade preferences", func() {
		It("should persist and reload account preferences", func() {
			prefs := snaptrade.Preferences{HiddenAccounts: []string{"acct-2"}, DefaultAccountID: "acct-1"}

			Expect(cli.SaveSnapTradePreferences(dep, prefs)).To(Succeed())

			out := cli.LoadSnapTradePreferences(dep)
			Expect(out.DefaultAccountID).To(Equal("acct-1"))
			Expect(out.IsHidden("acct-2")).To(BeTrue())
		})

		It("should return empty preferences when none are stored", func() {
			out := cli.LoadSnapTradePreferences(dep)

			Expect(out.HiddenAccounts).To(BeEmpty())
			Expect(out.DefaultAccountID).To(BeEmpty())
		})
	})

	Describe("ConnectSnapTrade", func() {

		When("required config is missing", func() {
			It("should return a helpful error naming the missing fields", func() {
				err := cli.ConnectSnapTrade(dep, c.Config{SnapTrade: c.ConfigSnapTrade{AccountType: "commercial"}}, func(string) error { return nil })

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("client-id"))
				Expect(err.Error()).To(ContainSubstring("consumer-key"))
				Expect(err.Error()).To(ContainSubstring("user-id"))
			})
		})

		When("using personal keys", func() {
			It("should skip registration and open the connection URL", func() {
				personalConfig := c.Config{SnapTrade: c.ConfigSnapTrade{ClientID: "client-id", ConsumerKey: "consumer-key"}}

				registered := false
				server.RouteToHandler("POST", "/snapTrade/registerUser", func(_ http.ResponseWriter, _ *http.Request) {
					registered = true
				})
				server.RouteToHandler("POST", "/snapTrade/login",
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]string{"redirectURI": "https://app.snaptrade.com/connect/xyz"}),
				)

				var opened string
				err := cli.ConnectSnapTrade(dep, personalConfig, func(uri string) error {
					opened = uri

					return nil
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(registered).To(BeFalse())
				Expect(opened).To(Equal("https://app.snaptrade.com/connect/xyz"))
			})
		})

		When("the user is not yet registered", func() {
			It("should register, persist the secret, and open the connection URL", func() {
				server.RouteToHandler("POST", "/snapTrade/registerUser",
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]string{"userId": "ben", "userSecret": "new-secret"}),
				)
				server.RouteToHandler("POST", "/snapTrade/login",
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]string{"redirectURI": "https://app.snaptrade.com/connect/abc"}),
				)

				var opened string
				err := cli.ConnectSnapTrade(dep, config, func(uri string) error {
					opened = uri

					return nil
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(opened).To(Equal("https://app.snaptrade.com/connect/abc"))

				secret, ok, _ := snaptrade.NewStore(fs, xdg.DataHome).GetUserSecret("ben")
				Expect(ok).To(BeTrue())
				Expect(secret).To(Equal("new-secret"))
			})
		})
	})
})

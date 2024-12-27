package cli_test

import (
	"errors"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	g "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/achannarasappa/ticker/v4/internal/cli"
	. "github.com/achannarasappa/ticker/v4/internal/cli"
	c "github.com/achannarasappa/ticker/v4/internal/common"
	httpMocks "github.com/achannarasappa/ticker/v4/test/http"
)

func getStdout(fn func()) string {
	rescueStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	out, _ := ioutil.ReadAll(r)
	os.Stdout = rescueStdout

	return string(out)
}

//nolint:errcheck
func writeConfigFile(fs afero.Fs, contents string) {
	home, _ := homedir.Dir()
	fs.MkdirAll(home, 0755)
	fs.Create(home + "/.ticker.yaml")
	afero.WriteFile(fs, home+"/.ticker.yaml", []byte(contents), 0644)

}

var _ = Describe("Cli", func() {

	var (
		options Options
		dep     c.Dependencies
	)

	BeforeEach(func() {
		options = Options{
			Watchlist:             "GME,BB",
			RefreshInterval:       0,
			Separate:              false,
			ExtraInfoExchange:     false,
			ExtraInfoFundamentals: false,
			ShowSummary:           false,
			ShowHoldings:          false,
			Proxy:                 "",
			Sort:                  "",
		}
		dep = c.Dependencies{
			Fs: afero.NewMemMapFs(),
			HttpClients: c.DependenciesHttpClients{
				Default:      client,
				Yahoo:        client,
				YahooSession: client,
			},
		}

		httpMocks.MockTickerSymbols()
		httpMocks.MockResponseCurrency()
		httpMocks.MockResponseForRefreshSessionSuccess()

		//nolint:errcheck
		dep.Fs.MkdirAll("./", 0755)
	})

	Describe("Run", func() {
		It("should start the UI", func() {
			fnIsCalled := false
			inputStartFn := func() error {
				fnIsCalled = true
				return nil
			}
			output := getStdout(func() {
				Run(inputStartFn)(&cobra.Command{}, []string{})
			})
			Expect(output).To(Equal(""))
			Expect(fnIsCalled).To(Equal(true))
		})

		When("the UI returns an error", func() {
			It("should report the error", func() {
				fnIsCalled := false
				inputStartFn := func() error {
					fnIsCalled = true
					return errors.New("ui error")
				}
				output := getStdout(func() {
					Run(inputStartFn)(&cobra.Command{}, []string{})
				})
				Expect(output).To(Equal("unable to start UI: ui error\n"))
				Expect(fnIsCalled).To(Equal(true))
			})
		})
	})

	Describe("GetContext", func() {

		Context("watchlist and groups", func() {

			type Case struct {
				InputOptions            cli.Options
				InputConfigFileContents string
				InputConfigFilePath     string
				AssertionErr            types.GomegaMatcher
				AssertionCtx            types.GomegaMatcher
			}

			DescribeTable("context values",
				func(c Case) {
					if c.InputConfigFileContents != "" {
						writeConfigFile(dep.Fs, c.InputConfigFileContents)
					}
					outputConfig, outputErr := cli.GetConfig(dep, c.InputConfigFilePath, c.InputOptions)
					outputCtx, outputErr := cli.GetContext(dep, outputConfig)
					Expect(outputErr).To(c.AssertionErr)
					Expect(outputCtx).To(c.AssertionCtx)
				},

				// option: watchlist
				Entry("when watchlist is set in config file", Case{
					InputOptions:            cli.Options{},
					InputConfigFileContents: "watchlist:\n  - GME\n  - BB",
					AssertionErr:            BeNil(),
					AssertionCtx: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Groups": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"ConfigAssetGroup": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Name":      Equal("default"),
									"Watchlist": Equal([]string{"GME", "BB"}),
								}),
							}),
						}),
					}),
				}),

				Entry("when watchlist is set in options", Case{
					InputOptions:            cli.Options{Watchlist: "BIO,BB"},
					InputConfigFileContents: "",
					AssertionErr:            BeNil(),
					AssertionCtx: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Groups": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"ConfigAssetGroup": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Name":      Equal("default"),
									"Watchlist": Equal([]string{"BIO", "BB"}),
								}),
							}),
						}),
					}),
				}),

				Entry("when watchlist is set in both config file and options", Case{
					InputOptions:            cli.Options{Watchlist: "BB"},
					InputConfigFileContents: "watchlist:\n  - GME",
					AssertionErr:            BeNil(),
					AssertionCtx: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Groups": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"ConfigAssetGroup": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Name":      Equal("default"),
									"Watchlist": Equal([]string{"BB"}),
								}),
							}),
						}),
					}),
				}),

				// groups
				Entry("when groups are defined", Case{
					InputOptions: cli.Options{},
					InputConfigFileContents: strings.Join([]string{
						"groups:",
						"  - name: crypto",
						"    watchlist:",
						"      - SHIB-USD",
						"      - BTC-USD",
						"    holdings:",
						"      - symbol: SOL1-USD",
						"        quantity: 17",
						"        unit_cost: 159.10",
					}, "\n"),
					AssertionErr: BeNil(),
					AssertionCtx: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Groups": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"ConfigAssetGroup": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Name":      Equal("crypto"),
									"Watchlist": Equal([]string{"SHIB-USD", "BTC-USD"}),
									"Holdings": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
										"0": g.MatchFields(g.IgnoreExtras, g.Fields{
											"Symbol":   Equal("SOL1-USD"),
											"Quantity": Equal(17.0),
											"UnitCost": Equal(159.10),
										}),
									}),
								}),
							}),
						}),
					}),
				}),

				Entry("when groups and watchlist are defined", Case{
					InputOptions: cli.Options{},
					InputConfigFileContents: strings.Join([]string{
						"watchlist:",
						"  - TSLA",
						"groups:",
						"  - name: crypto",
						"    watchlist:",
						"      - SOL1-USD",
						"      - BTC-USD",
					}, "\n"),
					AssertionErr: BeNil(),
					AssertionCtx: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Groups": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"ConfigAssetGroup": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Name":      Equal("default"),
									"Watchlist": Equal([]string{"TSLA"}),
								}),
							}),
							"1": g.MatchFields(g.IgnoreExtras, g.Fields{
								"ConfigAssetGroup": g.MatchFields(g.IgnoreExtras, g.Fields{
									"Name":      Equal("crypto"),
									"Watchlist": Equal([]string{"SOL1-USD", "BTC-USD"}),
								}),
							}),
						}),
					}),
				}),

				// symbols by source
				Entry("when groups and watchlist are defined", Case{
					InputOptions: cli.Options{},
					InputConfigFileContents: strings.Join([]string{
						"watchlist:",
						"  - TSLA",               // yahoo finance
						"  - ETHEREUM.CG",        // coingecko
						"  - BITCOIN.CC",         // coincap
						"  - ADA.CB",             // coinbase
						"  - BIT-31JAN25-CDE.CB", // coinbase futures
						"  - SOL.X",              // ticker
					}, "\n"),
					AssertionErr: BeNil(),
					AssertionCtx: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Groups": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
							"0": g.MatchFields(g.IgnoreExtras, g.Fields{
								"SymbolsBySource": g.MatchAllElements(func(element interface{}) string {
									return strconv.FormatInt(int64(element.(c.AssetGroupSymbolsBySource).Source), 10)
								}, g.Elements{
									"0": g.MatchFields(g.IgnoreExtras, g.Fields{
										"Symbols": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
											"0": Equal("TSLA"),
										}),
										"Source": Equal(c.QuoteSourceYahoo),
									}),
									"2": g.MatchFields(g.IgnoreExtras, g.Fields{
										"Symbols": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
											"0": Equal("ethereum"),
											"1": Equal("solana"),
										}),
										"Source": Equal(c.QuoteSourceCoingecko),
									}),
									"4": g.MatchFields(g.IgnoreExtras, g.Fields{
										"Symbols": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
											"0": Equal("bitcoin"),
										}),
										"Source": Equal(c.QuoteSourceCoinCap),
									}),
									"5": g.MatchFields(g.IgnoreExtras, g.Fields{
										"Symbols": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
											"0": Equal("ADA-USD"),
											"1": Equal("BIT-31JAN25-CDE"),
										}),
										"Source": Equal(c.QuoteSourceCoinbase),
									}),
								}),
							}),
						}),
					}),
				}),
			)

		})

		When("there is an error getting ticker symbols", func() {

			It("returns the error", func() {

				httpMocks.MockTickerSymbolsError()

				_, outputErr := GetContext(dep, c.Config{})

				Expect(outputErr).ToNot(BeNil())

			})

		})

		When("there is an error getting reference data", func() {

			It("returns the error", func() {

				httpMocks.MockResponseCurrencyError()

				_, outputErr := GetContext(dep, c.Config{
					Watchlist: []string{"TSLA"},
				})

				Expect(outputErr).ToNot(BeNil())

			})

		})

		When("there is an error refreshing the yahoo session", func() {

			It("returns the error", func() {

				httpMocks.MockResponseForRefreshSessionError()

				_, outputErr := GetContext(dep, c.Config{})

				Expect(outputErr).ToNot(BeNil())

			})

		})

	})

	Describe("GetConfig", func() {

		Context("options and configuration", func() {
			type Case struct {
				InputOptions            cli.Options
				InputConfigFileContents string
				InputConfigFilePath     string
				AssertionErr            types.GomegaMatcher
				AssertionConfig         types.GomegaMatcher
			}

			DescribeTable("config values",
				func(c Case) {
					if c.InputConfigFileContents != "" {
						writeConfigFile(dep.Fs, c.InputConfigFileContents)
					}
					outputConfig, outputErr := cli.GetConfig(dep, c.InputConfigFilePath, c.InputOptions)
					Expect(outputErr).To(c.AssertionErr)
					Expect(outputConfig).To(c.AssertionConfig)
				},

				// option: string (proxy, sort)
				Entry("when proxy is set in config file", Case{
					InputOptions:            cli.Options{},
					InputConfigFileContents: "proxy: http://myproxy.com:4438",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Proxy": Equal("http://myproxy.com:4438"),
					}),
				}),

				Entry("when proxy is set in options", Case{
					InputOptions:            cli.Options{Proxy: "http://www.example.org:3128"},
					InputConfigFileContents: "",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Proxy": Equal("http://www.example.org:3128"),
					}),
				}),

				Entry("when proxy is set in both config file and options", Case{
					InputOptions:            cli.Options{Proxy: "http://www.example.org:3128"},
					InputConfigFileContents: "proxy: http://myproxy.com:4438",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Proxy": Equal("http://www.example.org:3128"),
					}),
				}),

				// option: interval
				Entry("when interval is set in config file", Case{
					InputOptions:            cli.Options{},
					InputConfigFileContents: "interval: 8",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"RefreshInterval": Equal(8),
					}),
				}),

				Entry("when interval is set in options", Case{
					InputOptions:            cli.Options{RefreshInterval: 7},
					InputConfigFileContents: "",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"RefreshInterval": Equal(7),
					}),
				}),

				Entry("when interval is set in both config file and options", Case{
					InputOptions:            cli.Options{RefreshInterval: 7},
					InputConfigFileContents: "interval: 8",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"RefreshInterval": Equal(7),
					}),
				}),

				Entry("when interval is set in neither config file and options", Case{
					InputOptions:            cli.Options{},
					InputConfigFileContents: "",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"RefreshInterval": Equal(5),
					}),
				}),

				// option: boolean (separator, summary, fundamentals, tags, holdings)
				Entry("when show-separator is set in config file", Case{
					InputOptions:            cli.Options{},
					InputConfigFileContents: "show-separator: true",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Separate": Equal(true),
					}),
				}),

				Entry("when show-separator is set in options", Case{
					InputOptions:            cli.Options{Separate: true},
					InputConfigFileContents: "",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Separate": Equal(true),
					}),
				}),

				Entry("when show-separator is set in both config file and options", Case{
					InputOptions:            cli.Options{Separate: false},
					InputConfigFileContents: "show-separator: true",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Separate": Equal(true),
					}),
				}),
			)

		})

		//nolint:errcheck
		Context("reading the config file", func() {
			var (
				depLocal c.Dependencies
			)

			BeforeEach(func() {
				depLocal = c.Dependencies{
					Fs: afero.NewMemMapFs(),
					HttpClients: c.DependenciesHttpClients{
						Default: client,
						Yahoo:   client,
					},
				}
				afero.WriteFile(depLocal.Fs, ".ticker.yaml", []byte("watchlist:\n  - NOK"), 0644)
			})

			When("an explicit config file is set", func() {
				It("should read the config file from disk", func() {
					inputConfigPath := ".ticker.yaml"
					outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})

					Expect(outputConfig.Watchlist).To(Equal([]string{"NOK"}))
					Expect(outputErr).To(BeNil())
				})
			})

			When("the config path option is empty", func() {
				When("there is no config file on disk", func() {
					It("should return an empty config and no error", func() {
						inputHome, _ := homedir.Dir()
						inputConfigPath := ""
						depLocal.Fs.MkdirAll(inputHome, 0755)
						outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})

						Expect(outputErr).To(BeNil())
						Expect(outputConfig).To(Equal(c.Config{RefreshInterval: 5}))
					})
				})
				When("there is a config file in the home directory", func() {
					It("should read the config file from disk", func() {
						inputHome, _ := homedir.Dir()
						inputConfigPath := ""
						depLocal.Fs.MkdirAll(inputHome, 0755)
						depLocal.Fs.Create(inputHome + "/.ticker.yaml")
						afero.WriteFile(depLocal.Fs, inputHome+"/.ticker.yaml", []byte("watchlist:\n  - AMD"), 0644)
						outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})

						Expect(outputConfig.Watchlist).To(Equal([]string{"AMD"}))
						Expect(outputErr).To(BeNil())
					})
				})
				When("there is a config file in the current directory", func() {
					It("should read the config file from disk", func() {
						inputCurrentDirectory, _ := os.Getwd()
						inputConfigPath := ""
						depLocal.Fs.MkdirAll(inputCurrentDirectory, 0755)
						depLocal.Fs.Create(inputCurrentDirectory + "/.ticker.yaml")
						afero.WriteFile(depLocal.Fs, inputCurrentDirectory+"/.ticker.yaml", []byte("watchlist:\n  - JNJ"), 0644)
						outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})

						Expect(outputConfig.Watchlist).To(Equal([]string{"JNJ"}))
						Expect(outputErr).To(BeNil())
					})
				})
				When("there is a config file in the XDG config directory", func() {
					XIt("should read the config file from disk", func() {
						inputHome, _ := homedir.Dir()
						inputConfigHome := inputHome + "/.config"
						os.Setenv("XDG_CONFIG_HOME", inputConfigHome)
						inputConfigPath := ""
						depLocal.Fs.MkdirAll(inputConfigHome, 0755)
						depLocal.Fs.Create(inputConfigHome + "/.ticker.yaml")
						afero.WriteFile(depLocal.Fs, inputConfigHome+"/.ticker.yaml", []byte("watchlist:\n  - ABNB"), 0644)
						outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})
						os.Unsetenv("XDG_CONFIG_HOME")

						Expect(outputConfig.Watchlist).To(Equal([]string{"ABNB"}))
						Expect(outputErr).To(BeNil())
					})
				})
			})

			When("there is an error reading the config file", func() {
				It("returns the error", func() {
					inputConfigPath := ".config-file-that-does-not-exist.yaml"
					outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})

					Expect(outputConfig).To(Equal(c.Config{}))
					Expect(outputErr).To(MatchError("invalid config: open .config-file-that-does-not-exist.yaml: file does not exist"))
				})
			})

			When("there is an error parsing the config file", func() {
				It("returns the error", func() {
					inputConfigPath := ".ticker.yaml"
					afero.WriteFile(depLocal.Fs, ".ticker.yaml", []byte("watchlist:\n   NOK"), 0644)
					outputConfig, outputErr := GetConfig(depLocal, inputConfigPath, cli.Options{})

					Expect(outputConfig).To(Equal(c.Config{}))
					Expect(outputErr).To(MatchError("invalid config: yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `NOK` into []string"))

				})
			})
		})
	})

	Describe("GetDependencies", func() {

		It("should dependencies", func() {

			output := GetDependencies()
			expected := g.MatchAllFields(g.Fields{
				"Fs":          BeAssignableToTypeOf(afero.NewOsFs()),
				"HttpClients": BeAssignableToTypeOf(c.DependenciesHttpClients{}),
			})

			Expect(output).To(expected)

		})

	})

	Describe("Validate", func() {

		var (
			config c.Config
		)

		BeforeEach(func() {
			config = c.Config{}
		})

		When("a deferred error is passed in", func() {
			It("validation fails", func() {
				inputErr := errors.New("some config error")
				outputErr := Validate(&config, &cli.Options{}, &inputErr)(&cobra.Command{}, []string{})
				Expect(outputErr).To(MatchError("some config error"))
			})
		})

		Describe("watchlist", func() {
			When("there is no watchlist in the config file and no watchlist cli argument", func() {
				It("should return an error", func() {
					options.Watchlist = ""
					outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(MatchError("invalid config: No watchlist provided"))
				})

				When("a nil error reference is passed in from Cobra", func() {

					It("should not return an error", func() {
						var prevErr error
						outputErr := Validate(&config, &options, &prevErr)(&cobra.Command{}, []string{})
						Expect(outputErr).NotTo(HaveOccurred())
					})

				})

				When("there are lots set", func() {
					It("should not return an error", func() {
						options.Watchlist = ""
						config = c.Config{
							Lots: []c.Lot{
								{
									Symbol:   "SYM",
									UnitCost: 1.0,
									Quantity: 1.0,
								},
							},
						}
						outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
						Expect(outputErr).NotTo(HaveOccurred())
					})
				})
			})
		})

	})
})

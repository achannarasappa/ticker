package cli_test

import (
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/mitchellh/go-homedir"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	g "github.com/onsi/gomega/gstruct"
	"github.com/onsi/gomega/types"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"github.com/achannarasappa/ticker/v5/internal/cli"
	. "github.com/achannarasappa/ticker/v5/internal/cli"
	c "github.com/achannarasappa/ticker/v5/internal/common"
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
		server  *ghttp.Server
	)

	AfterEach(func() {
		server.Close()
	})

	BeforeEach(func() {

		server = ghttp.NewServer()

		options = Options{
			Watchlist:             "GME,BB",
			RefreshInterval:       0,
			Separate:              false,
			ExtraInfoExchange:     false,
			ExtraInfoFundamentals: false,
			ShowSummary:           false,
			ShowHoldings:          false,
			Sort:                  "",
		}
		dep = c.Dependencies{
			Fs:         afero.NewMemMapFs(),
			SymbolsURL: server.URL() + "/symbols.csv",
		}

		//nolint:errcheck
		dep.Fs.MkdirAll("./", 0755)

		// Mock the ticker symbols endpoint
		responseFixture := `"ADA.X","ADA-USD","cb"
"ALGO.X","ALGO-USD","cb"
"BTC.X","BTC-USD","cb"
"ETH.X","ETH-USD","cb"
"SOL.X","SOL-USD","cb"
"XRP.X","XRP-USD","cb"
`
		server.RouteToHandler("GET", "/symbols.csv",
			ghttp.CombineHandlers(
				ghttp.RespondWith(http.StatusOK, responseFixture, http.Header{"Content-Type": []string{"text/plain; charset=utf-8"}}),
			),
		)
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
									"5": g.MatchFields(g.IgnoreExtras, g.Fields{
										"Symbols": g.MatchAllElementsWithIndex(g.IndexIdentity, g.Elements{
											"0": Equal("ADA-USD"),
											"1": Equal("BIT-31JAN25-CDE"),
											"2": Equal("SOL-USD"),
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

				dep := c.Dependencies{
					Fs:         afero.NewMemMapFs(),
					SymbolsURL: "invalid-url",
				}

				_, outputErr := GetContext(dep, c.Config{})

				Expect(outputErr).ToNot(BeNil())

			})

		})

		When("there is an error getting the logger", func() {

			It("returns the error", func() {
				dep := c.Dependencies{
					Fs:         afero.NewMemMapFs(),
					SymbolsURL: server.URL() + "/symbols.csv",
				}

				// Create a read-only filesystem to force an error when trying to create the log file
				dep.Fs = afero.NewReadOnlyFs(dep.Fs)

				_, outputErr := GetContext(dep, c.Config{Debug: true})

				Expect(outputErr).To(MatchError("failed to create log file: operation not permitted"))
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

				// option: debug
				Entry("when debug is set in config file", Case{
					InputOptions:            cli.Options{},
					InputConfigFileContents: "debug: true",
					AssertionErr:            BeNil(),
					AssertionConfig: g.MatchFields(g.IgnoreExtras, g.Fields{
						"Debug": Equal(true),
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
			expected := g.MatchFields(g.IgnoreExtras, g.Fields{
				"Fs": BeAssignableToTypeOf(afero.NewOsFs()),
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

		Describe("currency", func() {
			When("a mixed-case currency is specified in the config file", func() {
				It("should return an error (even if a valid minor currency)", func() {
					options.Watchlist = "SEIT.L"
					config = c.Config{
						Currency: "USd",
					}
					outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(MatchError("invalid config: Display currency may only be an ISO 4217 major currency or blank (eg GBP not GBp; default: USD)"))
				})
			})

			When("a blank currency is specified in the config file", func() {
				It("should not return an error", func() {
					options.Watchlist = "SEIT.L"
					config := c.Config{
						Currency: "",
					}
					outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).NotTo(HaveOccurred())
				})
			})

			When("a short currency (len < 3) is specified in the config file", func() {
				It("should return an error", func() {
					options.Watchlist = "SEIT.L"
					config := c.Config{
						Currency: "US",
					}
					outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(MatchError("invalid config: Display currency may only be an ISO 4217 major currency or blank (eg GBP not GBp; default: USD)"))
				})
			})

			When("a long currency (len > 3) is specified in the config file", func() {
				It("should return an error", func() {
					options.Watchlist = "SEIT.L"
					config := c.Config{
						Currency: "USD2",
					}
					outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(MatchError("invalid config: Display currency may only be an ISO 4217 major currency or blank (eg GBP not GBp; default: USD)"))
				})
			})

			PWhen("a non-ISO 4217 currency is specified in the config file", func() {
				PIt("should return an error", func() {
					options.Watchlist = "SEIT.L"
					config = c.Config{
						Currency: "XXX",
					}
					outputErr := Validate(&config, &options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(MatchError("invalid config: Display currency may only be an ISO 4217 major currency or blank (eg GBP not GBp; default: USD)"))
				})
			})
		})

	})
})

package cli_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/mitchellh/go-homedir"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	"ticker/internal/cli"
	. "ticker/internal/cli"
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

var _ = Describe("Cli", func() {
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
				Expect(output).To(Equal("Unable to start UI: ui error\n"))
				Expect(fnIsCalled).To(Equal(true))
			})
		})
	})
	Describe("Validate", func() {

		var (
			options               cli.Options
			fs                    afero.Fs
			watchlist             string
			refreshInterval       int
			separate              bool
			extraInfoExchange     bool
			extraInfoFundamentals bool
			showTotals            bool
		)

		BeforeEach(func() {
			options = cli.Options{
				Watchlist:             &watchlist,
				RefreshInterval:       &refreshInterval,
				Separate:              &separate,
				ExtraInfoExchange:     &extraInfoExchange,
				ExtraInfoFundamentals: &extraInfoFundamentals,
				ShowTotals:            &showTotals,
			}
			watchlist = "GME,BB"
			refreshInterval = 0
			separate = false
			extraInfoExchange = false
			extraInfoFundamentals = false
			showTotals = false
			fs = afero.NewMemMapFs()
			fs.MkdirAll("./", 0755)
		})

		It("should set the config", func() {
			inputConfig := cli.Config{}
			expected := cli.Config{
				RefreshInterval: 5,
				Watchlist: []string{
					"GME",
					"BB",
				},
				Lots: nil,
			}
			outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
			Expect(outputErr).To(BeNil())
			Expect(inputConfig).To(Equal(expected))
		})

		When("a deferred error is passed in", func() {
			It("validation fails", func() {
				inputConfig := cli.Config{}
				outputErr := Validate(&inputConfig, fs, options, errors.New("some config error"))(&cobra.Command{}, []string{})
				Expect(outputErr).To(MatchError("some config error"))
			})
		})

		Describe("watchlist", func() {
			When("there is no watchlist in the config file and no watchlist cli argument", func() {
				It("should return an error", func() {
					watchlist = ""
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(MatchError("Invalid config: No watchlist provided"))

				})
			})

			When("there is a watchlist as a cli argument", func() {
				It("should set the watchlist from the cli argument", func() {
					watchlist = "AAPL,TW"
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.Watchlist).To(Equal([]string{
						"AAPL",
						"TW",
					}))
				})

				When("the config file also has a watchlist defined", func() {
					It("should set the watchlist from the cli argument", func() {
						watchlist = "F,GM"
						inputConfig := cli.Config{
							Watchlist: []string{"BIO"},
						}
						outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
						Expect(outputErr).To(BeNil())
						Expect(inputConfig.Watchlist).To(Equal([]string{
							"F",
							"GM",
						}))
					})
				})
			})

			When("there is a watchlist in the config file", func() {
				It("should set the watchlist from the config file", func() {
					watchlist = ""
					inputConfig := cli.Config{
						Watchlist: []string{"NET"},
					}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.Watchlist).To(Equal([]string{
						"NET",
					}))
				})
			})
		})

		Describe("refresh interval option", func() {
			When("refresh interval is set as a cli argument", func() {
				It("should set the config to the cli argument value", func() {
					refreshInterval = 9
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.RefreshInterval).To(Equal(9))
				})

				When("the config file also has a refresh interval defined", func() {
					It("should set the refresh interval from the cli argument", func() {
						refreshInterval = 8
						inputConfig := cli.Config{
							RefreshInterval: 7,
						}
						outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
						Expect(outputErr).To(BeNil())
						Expect(inputConfig.RefreshInterval).To(Equal(8))
					})
				})
			})

			When("refresh interval is set in the config file", func() {
				It("should set the config to the config argument value", func() {
					inputConfig := cli.Config{
						RefreshInterval: 357,
					}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.RefreshInterval).To(Equal(357))
				})
			})

			When("refresh interval is not set", func() {
				It("should set a default watch interval", func() {
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.RefreshInterval).To(Equal(5))
				})
			})
		})

		Describe("show-separator option", func() {
			When("show-separator flag is set as a cli argument", func() {
				It("should set the config to the cli argument value", func() {
					separate = true
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.Separate).To(Equal(true))
				})

				When("the config file also has a show-separator flag defined", func() {
					It("should set the show-separator flag from the cli argument", func() {
						separate = true
						inputConfig := cli.Config{
							Separate: true,
						}
						outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
						Expect(outputErr).To(BeNil())
						Expect(inputConfig.Separate).To(Equal(true))
					})
				})
			})

			When("show-separator flag is set in the config file", func() {
				It("should set the config to the config argument value", func() {
					inputConfig := cli.Config{
						Separate: true,
					}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.Separate).To(Equal(true))
				})
			})

			When("show-separator flag is not set", func() {
				It("should set a default watch interval", func() {
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.Separate).To(Equal(false))
				})
			})
		})

		Describe("show-tags option", func() {
			When("show-tags flag is set as a cli argument", func() {
				It("should set the config to the cli argument value", func() {
					extraInfoExchange = true
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.ExtraInfoExchange).To(Equal(true))
				})

				When("the config file also has a show-tags flag defined", func() {
					It("should set the show-tags flag from the cli argument", func() {
						extraInfoExchange = true
						inputConfig := cli.Config{
							ExtraInfoExchange: false,
						}
						outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
						Expect(outputErr).To(BeNil())
						Expect(inputConfig.ExtraInfoExchange).To(Equal(true))
					})
				})
			})

			When("show-tags flag is set in the config file", func() {
				It("should set the config to the config argument value", func() {
					inputConfig := cli.Config{
						ExtraInfoExchange: true,
					}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.ExtraInfoExchange).To(Equal(true))
				})
			})

			When("show-tags flag is not set", func() {
				It("should disable the option", func() {
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.ExtraInfoExchange).To(Equal(false))
				})
			})
		})

		Describe("show-fundamentals option", func() {
			When("show-fundamentals flag is set as a cli argument", func() {
				It("should set the config to the cli argument value", func() {
					extraInfoFundamentals = true
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.ExtraInfoFundamentals).To(Equal(true))
				})

				When("the config file also has a show-fundamentals flag defined", func() {
					It("should set the show-fundamentals flag from the cli argument", func() {
						extraInfoFundamentals = true
						inputConfig := cli.Config{
							ExtraInfoFundamentals: false,
						}
						outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
						Expect(outputErr).To(BeNil())
						Expect(inputConfig.ExtraInfoFundamentals).To(Equal(true))
					})
				})
			})

			When("show-fundamentals flag is set in the config file", func() {
				It("should set the config to the cli argument value", func() {
					inputConfig := cli.Config{
						ExtraInfoFundamentals: true,
					}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.ExtraInfoFundamentals).To(Equal(true))
				})
			})

			When("show-fundamentals flag is not set", func() {
				It("should disable the option", func() {
					inputConfig := cli.Config{}
					outputErr := Validate(&inputConfig, fs, options, nil)(&cobra.Command{}, []string{})
					Expect(outputErr).To(BeNil())
					Expect(inputConfig.ExtraInfoFundamentals).To(Equal(false))
				})
			})
		})
	})
	Describe("ReadConfig", func() {

		var (
			fs afero.Fs
		)

		BeforeEach(func() {
			fs = afero.NewMemMapFs()
			afero.WriteFile(fs, ".ticker.yaml", []byte("watchlist:\n  - NOK"), 0644)
		})

		When("an explicit config file is set", func() {
			It("should read the config file from disk", func() {
				inputConfigPath := ".ticker.yaml"
				config, err := ReadConfig(fs, inputConfigPath)

				Expect(config.Watchlist).To(Equal([]string{"NOK"}))
				Expect(err).To(BeNil())
			})
		})

		When("the config path option is empty", func() {
			When("there is no config file on disk", func() {
				It("should return an error", func() {
					inputHome, _ := homedir.Dir()
					inputConfigPath := ""
					fs.MkdirAll(inputHome, 0755)
					config, err := ReadConfig(fs, inputConfigPath)

					Expect(config).To(Equal(Config{}))
					Expect(err).ToNot(BeNil())
				})
			})
			When("there is a config file in the home directory", func() {
				It("should read the config file from disk", func() {
					inputHome, _ := homedir.Dir()
					inputConfigPath := ""
					fs.MkdirAll(inputHome, 0755)
					fs.Create(inputHome + "/.ticker.yaml")
					afero.WriteFile(fs, inputHome+"/.ticker.yaml", []byte("watchlist:\n  - AMD"), 0644)
					config, err := ReadConfig(fs, inputConfigPath)

					Expect(config.Watchlist).To(Equal([]string{"AMD"}))
					Expect(err).To(BeNil())
				})
			})
			When("there is a config file in the current directory", func() {
				It("should read the config file from disk", func() {
					inputCurrentDirectory, _ := os.Getwd()
					inputConfigPath := ""
					fs.MkdirAll(inputCurrentDirectory, 0755)
					fs.Create(inputCurrentDirectory + "/.ticker.yaml")
					afero.WriteFile(fs, inputCurrentDirectory+"/.ticker.yaml", []byte("watchlist:\n  - JNJ"), 0644)
					config, err := ReadConfig(fs, inputConfigPath)

					Expect(config.Watchlist).To(Equal([]string{"JNJ"}))
					Expect(err).To(BeNil())
				})
			})
			When("there is a config file in the XDG config directory", func() {
				It("should read the config file from disk", func() {
					inputHome, _ := homedir.Dir()
					inputConfigHome := inputHome + "/.config"
					os.Setenv("XDG_CONFIG_HOME", inputConfigHome)
					inputConfigPath := ""
					fs.MkdirAll(inputConfigHome, 0755)
					fs.Create(inputConfigHome + "/.ticker.yaml")
					afero.WriteFile(fs, inputConfigHome+"/.ticker.yaml", []byte("watchlist:\n  - ABNB"), 0644)
					config, err := ReadConfig(fs, inputConfigPath)
					os.Unsetenv("XDG_CONFIG_HOME")

					Expect(config.Watchlist).To(Equal([]string{"ABNB"}))
					Expect(err).To(BeNil())
				})
			})
		})

		When("there is an error reading the config file", func() {
			It("returns the error", func() {
				inputConfigPath := ".config-file-that-does-not-exist.yaml"
				config, err := ReadConfig(fs, inputConfigPath)

				Expect(config).To(Equal(cli.Config{}))
				Expect(err).To(MatchError("Invalid config: open .config-file-that-does-not-exist.yaml: file does not exist"))
			})
		})

		When("there is an error parsing the config file", func() {
			It("returns the error", func() {
				inputConfigPath := ".ticker.yaml"
				afero.WriteFile(fs, ".ticker.yaml", []byte("watchlist:\n   NOK"), 0644)
				config, err := ReadConfig(fs, inputConfigPath)

				Expect(config).To(Equal(cli.Config{}))
				Expect(err).To(MatchError("Invalid config: yaml: unmarshal errors:\n  line 2: cannot unmarshal !!str `NOK` into []string"))

			})
		})
	})
})

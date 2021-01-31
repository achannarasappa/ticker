package cli_test

import (
	"errors"
	"io/ioutil"
	"os"

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
			config          cli.Config
			options         cli.Options
			fs              afero.Fs
			watchlist       string
			refreshInterval int
			configPath      string
			compact         bool
		)

		BeforeEach(func() {
			config = cli.Config{}
			options = cli.Options{
				ConfigPath:      &configPath,
				Watchlist:       &watchlist,
				RefreshInterval: &refreshInterval,
				Compact:         &compact,
			}
			watchlist = "GME,BB"
			refreshInterval = 0
			compact = false
			configPath = ""
			fs = afero.NewMemMapFs()
			fs.MkdirAll("./", 0755)
		})

		It("should set the config", func() {
			expected := cli.Config{
				RefreshInterval: 5,
				Watchlist: []string{
					"GME",
					"BB",
				},
				Lots: nil,
			}
			output := Validate(&config, fs, options)(&cobra.Command{}, []string{})
			Expect(output).To(BeNil())
			Expect(config).To(Equal(expected))
		})

		When("there is a error opening the config file", func() {
			It("should return an error", func() {
				configPath = ".config-file-that-does-not-exist.yaml"
				output := Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(output).To(MatchError("Invalid config: open .config-file-that-does-not-exist.yaml: file does not exist"))
			})
		})

		When("there is a error parsing the config file", func() {
			It("should return an error", func() {
				configPath = ".ticker.yaml"
				afero.WriteFile(fs, ".ticker.yaml", []byte("invalid yaml content"), 0644)
				output := Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(output).To(MatchError("Invalid config: yaml: unmarshal errors:\n  line 1: cannot unmarshal !!str `invalid...` into cli.Config"))
			})
		})

		When("there is no watchlist in the config file and no watchlist cli argument", func() {
			It("should return an error", func() {
				watchlist = ""
				output := Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(output).To(MatchError("Invalid config: No watchlist provided"))

			})
		})

		When("there is a watchlist as a cli argument", func() {
			It("should set the watchlist from the cli argument", func() {
				watchlist = "AAPL,TW"
				Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(config.Watchlist).To(Equal([]string{
					"AAPL",
					"TW",
				}))
			})

			When("the config file also has a watchlist defined", func() {
				It("should set the watchlist from the cli argument", func() {
					watchlist = "F,GM"
					afero.WriteFile(fs, ".ticker.yaml", []byte("watchlist:\n  - BIO"), 0644)
					Validate(&config, fs, options)(&cobra.Command{}, []string{})
					Expect(config.Watchlist).To(Equal([]string{
						"F",
						"GM",
					}))
				})
			})
		})

		When("there is a watchlist in the config file", func() {
			It("should set the watchlist from the config file", func() {
				watchlist = ""
				configPath = ".ticker.yaml"
				afero.WriteFile(fs, ".ticker.yaml", []byte("watchlist:\n  - NET"), 0644)
				Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(config.Watchlist).To(Equal([]string{
					"NET",
				}))
			})
		})

		When("refresh interval is set as a cli argument", func() {
			It("should set the config to the cli argument value", func() {
				refreshInterval = 9
				Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(config.RefreshInterval).To(Equal(9))
			})

			When("the config file also has a refresh interval defined", func() {
				It("should set the refresh interval from the cli argument", func() {
					refreshInterval = 8
					config.RefreshInterval = 7
					Validate(&config, fs, options)(&cobra.Command{}, []string{})
					Expect(config.RefreshInterval).To(Equal(8))
				})
			})
		})

		When("refresh interval is set in the config file", func() {
			It("should set the config to the cli argument value", func() {
				config.RefreshInterval = 357
				Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(config.RefreshInterval).To(Equal(357))
			})
		})

		When("refresh interval is not set", func() {
			It("should set a default watch interval", func() {
				Validate(&config, fs, options)(&cobra.Command{}, []string{})
				Expect(config.RefreshInterval).To(Equal(5))
			})
		})

		Describe("compact option", func() {
			When("compact flag is set as a cli argument", func() {
				It("should set the config to the cli argument value", func() {
					compact = true
					Validate(&config, fs, options)(&cobra.Command{}, []string{})
					Expect(config.Compact).To(Equal(true))
				})

				When("the config file also has a compact flag defined", func() {
					It("should set the compact flag from the cli argument", func() {
						compact = true
						config.Compact = false
						Validate(&config, fs, options)(&cobra.Command{}, []string{})
						Expect(config.Compact).To(Equal(true))
					})
				})
			})

			When("compact flag is set in the config file", func() {
				It("should set the config to the cli argument value", func() {
					config.Compact = true
					Validate(&config, fs, options)(&cobra.Command{}, []string{})
					Expect(config.Compact).To(Equal(true))
				})
			})

			When("compact flag is not set", func() {
				It("should set a default watch interval", func() {
					Validate(&config, fs, options)(&cobra.Command{}, []string{})
					Expect(config.Compact).To(Equal(false))
				})
			})
		})
	})
})

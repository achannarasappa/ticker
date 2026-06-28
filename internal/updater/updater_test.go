package updater_test

import (
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/spf13/afero"

	"github.com/achannarasappa/ticker/v5/internal/cache"
	"github.com/achannarasappa/ticker/v5/internal/updater"
)

const (
	cacheFilePath  = "/cache/ticker/cache.json"
	currentVersion = "v5.0.0"
	latestVersion  = "v5.1.0"
)

// setupCache writes a cached latest-version entry with the given time-to-live so
// tests can simulate a fresh (positive ttl) or stale (negative ttl) cache.
func setupCache(fs afero.Fs, version string, ttl time.Duration) {
	cache.New(fs, cacheFilePath, true).Set("latest-version", version, ttl)
}

var _ = Describe("Check", func() {

	var (
		fs     afero.Fs
		server *ghttp.Server
	)

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
		server = ghttp.NewServer()
		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest("GET", "/releases/latest"),
				ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]string{"tag_name": latestVersion}),
			),
		)
	})

	AfterEach(func() {
		server.Close()
	})

	When("the current version is a dev build", func() {
		It("should return empty without making a network request", func() {
			output := updater.Check("v0.0.0", server.URL()+"/releases/latest", cacheFilePath, fs)
			Expect(output).To(BeEmpty())
			Expect(server.ReceivedRequests()).To(BeEmpty())
		})
	})

	When("the cache is fresh", func() {
		When("the cached version is newer than the current version", func() {
			It("should return the cached version without making a network request", func() {
				setupCache(fs, latestVersion, time.Hour)
				output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
				Expect(output).To(Equal(latestVersion))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})

		When("the cached version matches the current version", func() {
			It("should return empty without making a network request", func() {
				setupCache(fs, currentVersion, time.Hour)
				output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
				Expect(output).To(BeEmpty())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})
	})

	When("the cache is stale", func() {
		It("should fetch from the API and return the latest version", func() {
			setupCache(fs, currentVersion, -time.Minute)
			output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
			Expect(output).To(Equal(latestVersion))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})
	})

	When("the cache does not exist", func() {
		It("should fetch from the API and return the latest version", func() {
			output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
			Expect(output).To(Equal(latestVersion))
			Expect(server.ReceivedRequests()).To(HaveLen(1))
		})

		It("should write the cache after fetching", func() {
			updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
			exists, _ := afero.Exists(fs, cacheFilePath)
			Expect(exists).To(BeTrue())
		})
	})

	When("the API request fails", func() {
		It("should return empty silently", func() {
			output := updater.Check(currentVersion, "http://localhost:0/invalid", cacheFilePath, fs)
			Expect(output).To(BeEmpty())
		})
	})

	When("the latest version matches the current version", func() {
		It("should return empty", func() {
			server.SetHandler(0,
				ghttp.CombineHandlers(
					ghttp.RespondWithJSONEncoded(http.StatusOK, map[string]string{"tag_name": currentVersion}),
				),
			)
			output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
			Expect(output).To(BeEmpty())
		})
	})
})

package updater_test

import (
	"encoding/json"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/spf13/afero"

	"github.com/achannarasappa/ticker/v5/internal/updater"
)

const (
	cacheFilePath  = "/cache/ticker/update-check.json"
	currentVersion = "v5.0.0"
	latestVersion  = "v5.1.0"
)

func writeCacheFile(fs afero.Fs, checkedAt time.Time, version string) {
	type cacheEntry struct {
		CheckedAt     time.Time `json:"checked_at"`
		LatestVersion string    `json:"latest_version"`
	}
	data, _ := json.Marshal(cacheEntry{CheckedAt: checkedAt, LatestVersion: version})
	fs.MkdirAll("/cache/ticker", 0755) //nolint:errcheck
	afero.WriteFile(fs, cacheFilePath, data, 0644) //nolint:errcheck
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
				writeCacheFile(fs, time.Now(), latestVersion)
				output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
				Expect(output).To(Equal(latestVersion))
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})

		When("the cached version matches the current version", func() {
			It("should return empty without making a network request", func() {
				writeCacheFile(fs, time.Now(), currentVersion)
				output := updater.Check(currentVersion, server.URL()+"/releases/latest", cacheFilePath, fs)
				Expect(output).To(BeEmpty())
				Expect(server.ReceivedRequests()).To(BeEmpty())
			})
		})
	})

	When("the cache is stale", func() {
		It("should fetch from the API and return the latest version", func() {
			writeCacheFile(fs, time.Now().Add(-4*time.Hour), currentVersion)
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

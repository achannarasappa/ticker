package cache_test

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/afero"

	"github.com/achannarasappa/ticker/v5/internal/cache"
)

const (
	cachePath = "/cache/ticker/cache.json"
	testTTL   = time.Hour
)

type payload struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

var _ = Describe("Cache", func() {

	var fs afero.Fs

	BeforeEach(func() {
		fs = afero.NewMemMapFs()
	})

	Describe("FilePath", func() {
		It("returns a path under the ticker cache directory", func() {
			Expect(cache.FilePath()).To(HaveSuffix("ticker/cache.json"))
		})
	})

	Describe("Get and Set", func() {

		When("the cache is enabled", func() {

			It("round-trips a stored value", func() {
				cache := cache.New(fs, cachePath, true)
				cache.Set("key", payload{Name: "abc", Count: 3}, testTTL)

				var out payload
				hit := cache.Get("key", &out)

				Expect(hit).To(BeTrue())
				Expect(out).To(Equal(payload{Name: "abc", Count: 3}))
			})

			It("returns false for a key that was never set", func() {
				cache := cache.New(fs, cachePath, true)

				var out payload
				Expect(cache.Get("missing", &out)).To(BeFalse())
			})

			It("returns false for an entry older than its TTL", func() {
				cache := cache.New(fs, cachePath, true)
				cache.Set("key", payload{Name: "stale"}, -time.Minute)

				var out payload
				Expect(cache.Get("key", &out)).To(BeFalse())
			})
		})

		When("the cache is disabled", func() {

			It("does not store values and always misses", func() {
				cache := cache.New(fs, cachePath, false)
				cache.Set("key", payload{Name: "abc"}, testTTL)

				var out payload
				Expect(cache.Get("key", &out)).To(BeFalse())
			})

			It("does not write a cache file", func() {
				cache := cache.New(fs, cachePath, false)
				cache.Set("key", payload{Name: "abc"}, testTTL)

				exists, _ := afero.Exists(fs, cachePath)
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("persistence across instances", func() {

		It("reads values written by a previous instance sharing the same file", func() {
			writer := cache.New(fs, cachePath, true)
			writer.Set("key", payload{Name: "persisted", Count: 7}, testTTL)

			reader := cache.New(fs, cachePath, true)
			var out payload
			hit := reader.Get("key", &out)

			Expect(hit).To(BeTrue())
			Expect(out).To(Equal(payload{Name: "persisted", Count: 7}))
		})

		It("merges entries written by separate instances rather than clobbering", func() {
			a := cache.New(fs, cachePath, true)
			a.Set("a", payload{Name: "first"}, testTTL)

			b := cache.New(fs, cachePath, true)
			b.Set("b", payload{Name: "second"}, testTTL)

			reader := cache.New(fs, cachePath, true)
			var outA, outB payload
			Expect(reader.Get("a", &outA)).To(BeTrue())
			Expect(reader.Get("b", &outB)).To(BeTrue())
			Expect(outA.Name).To(Equal("first"))
			Expect(outB.Name).To(Equal("second"))
		})
	})

	Describe("cache schema versioning", func() {

		// writeRawEntry writes a cache file containing a single entry under the
		// given raw (already-namespaced) key, simulating data written by a
		// ticker version using a different cache schema.
		writeRawEntry := func(rawKey string, p payload) {
			type rawEntry struct {
				ExpiresAt time.Time       `json:"expires_at"`
				Payload   json.RawMessage `json:"payload"`
			}
			encodedPayload, _ := json.Marshal(p)
			data, _ := json.Marshal(map[string]rawEntry{
				rawKey: {ExpiresAt: time.Now().Add(time.Hour), Payload: encodedPayload},
			})
			afero.WriteFile(fs, cachePath, data, 0600) //nolint:errcheck
		}

		It("ignores entries written under a different schema version", func() {
			writeRawEntry("v0:key", payload{Name: "old"})

			var out payload
			Expect(cache.New(fs, cachePath, true).Get("key", &out)).To(BeFalse())
		})

		It("prunes entries from other schema versions when writing", func() {
			writeRawEntry("v0:key", payload{Name: "old"})

			cache.New(fs, cachePath, true).Set("key", payload{Name: "new"}, testTTL)

			var onDisk map[string]json.RawMessage
			data, _ := afero.ReadFile(fs, cachePath)
			Expect(json.Unmarshal(data, &onDisk)).To(Succeed())
			Expect(onDisk).NotTo(HaveKey("v0:key"))
			Expect(onDisk).To(HaveLen(1))
		})
	})
})

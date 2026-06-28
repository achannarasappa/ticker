// Package cache provides a simple file-based, inter-process cache for data
// fetched at startup (e.g. the symbol map, Yahoo session, currency lookups) and
// other slow-changing data such as the latest-version check. Multiple ticker
// instances running on the same machine share a single JSON file so that only
// the first instance to fetch a given piece of data (within its TTL) pays the
// cost of the network request.
//
// Each entry carries its own expiry, set by the caller via Set, so that
// long-lived data (symbol map, session, per-symbol currency) can be reused for
// much longer than volatile data (currency rates).
//
// All keys are namespaced with schemaVersion so that, after a ticker upgrade
// that changes the shape of any cached value, entries written by the previous
// version are simply not found (and are pruned on the next write) rather than
// being decoded into the new shape and silently misread. Bump schemaVersion
// whenever a cached value's structure changes.
package cache

import (
	"encoding/json"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"github.com/spf13/afero"
)

const schemaVersion = 1

type entry struct {
	ExpiresAt time.Time       `json:"expires_at"`
	Payload   json.RawMessage `json:"payload"`
}

// keyPrefix is prepended to every key to namespace entries by cache schema
// version (e.g. "v1:").
func keyPrefix() string {
	return "v" + strconv.Itoa(schemaVersion) + ":"
}

// Cache is a file-backed key/value store with a per-entry TTL. When disabled,
// all operations are no-ops and Get always reports a miss.
type Cache struct {
	fs      afero.Fs
	path    string
	enabled bool
	mu      sync.Mutex
	entries map[string]entry
}

// FilePath returns the default location of the shared cache file.
func FilePath() string {
	return filepath.Join(xdg.CacheHome, "ticker", "cache.json")
}

// New creates a cache backed by the file at path. When enabled is false the
// returned cache neither reads nor writes the file. The existing file (if any)
// is loaded into memory so reads during this session are served without
// re-reading the file.
func New(fs afero.Fs, path string, enabled bool) *Cache {
	cache := &Cache{
		fs:      fs,
		path:    path,
		enabled: enabled,
		entries: make(map[string]entry),
	}

	if enabled {
		cache.entries = readEntries(fs, path)
	}

	return cache
}

// Get reports whether key has a fresh entry and, if so, decodes its payload
// into out. It returns false on a disabled cache, a missing key, an expired
// entry, or a decode error.
func (c *Cache) Get(key string, out any) bool {
	if !c.enabled {
		return false
	}

	c.mu.Lock()
	cached, exists := c.entries[keyPrefix()+key]
	c.mu.Unlock()

	if !exists {
		return false
	}

	if !time.Now().Before(cached.ExpiresAt) {
		return false
	}

	if err := json.Unmarshal(cached.Payload, out); err != nil {
		return false
	}

	return true
}

// Set stores value under key with the given time-to-live and persists the cache
// file. It is a no-op on a disabled cache. Failures to marshal or persist are
// silently ignored - the cache is an optimization and must never break startup.
func (c *Cache) Set(key string, value any, ttl time.Duration) {
	if !c.enabled {
		return
	}

	payload, err := json.Marshal(value)
	if err != nil {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Re-read the file and merge so entries written by other instances since
	// this cache was created are not clobbered (last-writer-wins per key).
	c.entries = readEntries(c.fs, c.path)

	// Drop entries written under a different cache schema version so the file
	// does not accumulate unreadable data across ticker upgrades.
	prefix := keyPrefix()
	for existingKey := range c.entries {
		if !strings.HasPrefix(existingKey, prefix) {
			delete(c.entries, existingKey)
		}
	}

	c.entries[prefix+key] = entry{ExpiresAt: time.Now().Add(ttl), Payload: payload}

	writeEntries(c.fs, c.path, c.entries)
}

func readEntries(fs afero.Fs, path string) map[string]entry {
	entries := make(map[string]entry)

	data, err := afero.ReadFile(fs, path)
	if err != nil {
		return entries
	}

	if err := json.Unmarshal(data, &entries); err != nil {
		return make(map[string]entry)
	}

	return entries
}

func writeEntries(fs afero.Fs, path string, entries map[string]entry) {
	data, err := json.Marshal(entries)
	if err != nil {
		return
	}

	if err := fs.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return
	}

	// Write to a temp file and rename so a concurrent reader never observes a
	// partially written file.
	tmpPath := path + ".tmp"
	if err := afero.WriteFile(fs, tmpPath, data, 0600); err != nil {
		return
	}

	_ = fs.Rename(tmpPath, path)
}

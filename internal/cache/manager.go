// Package cache provides caching functionality for the GW2 MCP server.
package cache

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

// Manager handles caching for the GW2 MCP server
type Manager struct {
	cache *cache.Cache
}

// Key represents different types of cache keys
type Key string

const (
	// CurrencyListKey is the cache key for the list of all currencies
	CurrencyListKey Key = "currencies:list"
	// CurrencyDetailKey is the cache key template for individual currency details
	CurrencyDetailKey Key = "currency:detail:%d"
	// WikiSearchKey is the cache key template for wiki search results
	WikiSearchKey Key = "wiki:search:%s"
	// WikiPageKey is the cache key template for wiki page content
	WikiPageKey Key = "wiki:page:%s"

	// WalletKey is the cache key template for wallet data (short TTL)
	WalletKey Key = "wallet:%s" // %s = hashed API key
)

// Cache durations
const (
	// Static data - cache for very long periods
	StaticDataTTL = 24 * time.Hour * 365 // 1 year for currencies
	WikiDataTTL   = 24 * time.Hour       // 1 day for wiki content

	// Dynamic data - shorter cache periods
	WalletDataTTL = 5 * time.Minute // 5 minutes for wallet data

	// Default cleanup interval
	CleanupInterval = 10 * time.Minute
)

// NewManager creates a new cache manager
func NewManager() *Manager {
	return &Manager{
		cache: cache.New(StaticDataTTL, CleanupInterval),
	}
}

// Set stores a value in the cache with the specified TTL
func (m *Manager) Set(key string, value interface{}, ttl time.Duration) {
	m.cache.Set(key, value, ttl)
}

// Get retrieves a value from the cache
func (m *Manager) Get(key string) (interface{}, bool) {
	return m.cache.Get(key)
}

// GetString retrieves a string value from the cache
func (m *Manager) GetString(key string) (string, bool) {
	if value, found := m.cache.Get(key); found {
		if str, ok := value.(string); ok {
			return str, true
		}
	}
	return "", false
}

// GetJSON retrieves and unmarshals a JSON value from the cache
func (m *Manager) GetJSON(key string, dest interface{}) bool {
	if value, found := m.cache.Get(key); found {
		if jsonStr, ok := value.(string); ok {
			if err := json.Unmarshal([]byte(jsonStr), dest); err == nil {
				return true
			}
		}
	}
	return false
}

// SetJSON marshals and stores a JSON value in the cache
func (m *Manager) SetJSON(key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	m.cache.Set(key, string(jsonData), ttl)
	return nil
}

// Delete removes a value from the cache
func (m *Manager) Delete(key string) {
	m.cache.Delete(key)
}

// Flush clears all cached data
func (m *Manager) Flush() {
	m.cache.Flush()
}

// ItemCount returns the number of items in the cache
func (m *Manager) ItemCount() int {
	return m.cache.ItemCount()
}

// GetCurrencyListKey returns the cache key for currency list
func (m *Manager) GetCurrencyListKey() string {
	return string(CurrencyListKey)
}

// GetCurrencyDetailKey returns the cache key for a specific currency
func (m *Manager) GetCurrencyDetailKey(id int) string {
	return fmt.Sprintf(string(CurrencyDetailKey), id)
}

// GetWikiSearchKey returns the cache key for wiki search results
func (m *Manager) GetWikiSearchKey(query string) string {
	return fmt.Sprintf(string(WikiSearchKey), query)
}

// GetWikiPageKey returns the cache key for a wiki page
func (m *Manager) GetWikiPageKey(title string) string {
	return fmt.Sprintf(string(WikiPageKey), title)
}

// GetWalletKey returns the cache key for wallet data
func (m *Manager) GetWalletKey(apiKeyHash string) string {
	return fmt.Sprintf(string(WalletKey), apiKeyHash)
}

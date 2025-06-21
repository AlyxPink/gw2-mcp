// Package gw2api provides functionality for interacting with the Guild Wars 2 API.
package gw2api

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/log"

	"github.com/AlyxPink/gw2-mcp/internal/cache"
)

const (
	baseURL        = "https://api.guildwars2.com/v2"
	userAgent      = "github.com/AlyxPink/gw2-mcp"
	requestTimeout = 30 * time.Second
)

// Client handles GW2 API requests
type Client struct {
	httpClient *http.Client
	cache      *cache.Manager
	logger     *log.Logger
}

// WalletEntry represents a single currency in the wallet
type WalletEntry struct {
	ID    int `json:"id"`
	Value int `json:"value"`
}

// Currency represents currency metadata
type Currency struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	ID          int    `json:"id"`
	Order       int    `json:"order"`
}

// WalletInfo combines wallet entries with currency metadata
type WalletInfo struct {
	UpdatedAt  time.Time        `json:"updated_at"`
	Currencies map[int]Currency `json:"currencies"`
	Entries    []WalletEntry    `json:"entries"`
	Total      int              `json:"total_currencies"`
}

// NewClient creates a new GW2 API client
func NewClient(cacheManager *cache.Manager, logger *log.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		cache:  cacheManager,
		logger: logger,
	}
}

// GetWallet retrieves wallet information for the given API key
func (c *Client) GetWallet(ctx context.Context, apiKey string) (*WalletInfo, error) {
	// Create a hash of the API key for caching (security)
	hash := sha256.Sum256([]byte(apiKey))
	apiKeyHash := fmt.Sprintf("%x", hash[:8]) // Use first 8 bytes of hash

	cacheKey := c.cache.GetWalletKey(apiKeyHash)

	// Try to get from cache first
	var walletInfo WalletInfo
	if c.cache.GetJSON(cacheKey, &walletInfo) {
		c.logger.Debug("Wallet cache hit", "api_key_hash", apiKeyHash)
		return &walletInfo, nil
	}

	c.logger.Debug("Wallet cache miss, fetching from API", "api_key_hash", apiKeyHash)

	// Fetch wallet data from API
	walletEntries, err := c.fetchWallet(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch wallet: %w", err)
	}

	// Get currency metadata for all currencies in wallet
	currencyIDs := make([]int, len(walletEntries))
	for i, entry := range walletEntries {
		currencyIDs[i] = entry.ID
	}

	currencies, err := c.GetCurrencies(ctx, currencyIDs)
	if err != nil {
		c.logger.Warn("Failed to get currency metadata", "error", err)
		// Continue without metadata
		currencies = make(map[int]Currency)
	}

	// Create wallet info
	walletInfo = WalletInfo{
		Entries:    walletEntries,
		Currencies: currencies,
		Total:      len(walletEntries),
		UpdatedAt:  time.Now(),
	}

	// Cache the result
	if err := c.cache.SetJSON(cacheKey, walletInfo, cache.WalletDataTTL); err != nil {
		c.logger.Warn("Failed to cache wallet data", "error", err)
	}

	return &walletInfo, nil
}

// GetCurrencies retrieves currency metadata
func (c *Client) GetCurrencies(ctx context.Context, ids []int) (map[int]Currency, error) {
	// If no specific IDs requested, get all currencies
	if len(ids) == 0 {
		return c.getAllCurrencies(ctx)
	}

	// Get specific currencies
	currencies := make(map[int]Currency)
	var missingIDs []int

	// Check cache for each currency
	for _, id := range ids {
		cacheKey := c.cache.GetCurrencyDetailKey(id)
		var currency Currency
		if c.cache.GetJSON(cacheKey, &currency) {
			currencies[id] = currency
		} else {
			missingIDs = append(missingIDs, id)
		}
	}

	// Fetch missing currencies from API
	if len(missingIDs) > 0 {
		fetchedCurrencies, err := c.fetchCurrencies(ctx, missingIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch currencies: %w", err)
		}

		// Add fetched currencies to result and cache
		for _, currency := range fetchedCurrencies {
			currencies[currency.ID] = currency
			cacheKey := c.cache.GetCurrencyDetailKey(currency.ID)
			if err := c.cache.SetJSON(cacheKey, currency, cache.StaticDataTTL); err != nil {
				c.logger.Warn("Failed to cache currency", "id", currency.ID, "error", err)
			}
		}
	}

	return currencies, nil
}

// getAllCurrencies retrieves all available currencies
func (c *Client) getAllCurrencies(ctx context.Context) (map[int]Currency, error) {
	cacheKey := c.cache.GetCurrencyListKey()

	// Try cache first
	var currencies map[int]Currency
	if c.cache.GetJSON(cacheKey, &currencies) {
		c.logger.Debug("Currency list cache hit")
		return currencies, nil
	}

	c.logger.Debug("Currency list cache miss, fetching from API")

	// Fetch all currency IDs first
	currencyIDs, err := c.fetchCurrencyIDs(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch currency IDs: %w", err)
	}

	// Fetch all currency details
	currencyList, err := c.fetchCurrencies(ctx, currencyIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch currency details: %w", err)
	}

	// Convert to map
	currencies = make(map[int]Currency)
	for _, currency := range currencyList {
		currencies[currency.ID] = currency
	}

	// Cache the result
	if err := c.cache.SetJSON(cacheKey, currencies, cache.StaticDataTTL); err != nil {
		c.logger.Warn("Failed to cache currency list", "error", err)
	}

	return currencies, nil
}

// fetchWallet makes the actual API call to get wallet data
func (c *Client) fetchWallet(ctx context.Context, apiKey string) ([]WalletEntry, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/account/wallet", http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("API request failed with status %d and failed to read body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var wallet []WalletEntry
	if err := json.NewDecoder(resp.Body).Decode(&wallet); err != nil {
		return nil, err
	}

	return wallet, nil
}

// fetchCurrencyIDs fetches all available currency IDs
func (c *Client) fetchCurrencyIDs(ctx context.Context) ([]int, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/currencies", http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("API request failed with status %d and failed to read body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var ids []int
	if err := json.NewDecoder(resp.Body).Decode(&ids); err != nil {
		return nil, err
	}

	return ids, nil
}

// fetchCurrencies fetches currency details for specific IDs
func (c *Client) fetchCurrencies(ctx context.Context, ids []int) ([]Currency, error) {
	// Convert IDs to comma-separated string
	idStrs := make([]string, len(ids))
	for i, id := range ids {
		idStrs[i] = strconv.Itoa(id)
	}
	idsParam := strings.Join(idStrs, ",")

	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/currencies?ids="+idsParam, http.NoBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			c.logger.Warn("Failed to close response body", "error", closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return nil, fmt.Errorf("API request failed with status %d and failed to read body: %w", resp.StatusCode, readErr)
		}
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var currencies []Currency
	if err := json.NewDecoder(resp.Body).Decode(&currencies); err != nil {
		return nil, err
	}

	return currencies, nil
}

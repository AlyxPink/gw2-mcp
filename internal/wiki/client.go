package wiki

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AlyxPink/gw2-mcp/internal/cache"
	"github.com/charmbracelet/log"
)

const (
	wikiBaseURL    = "https://wiki.guildwars2.com"
	wikiAPIURL     = wikiBaseURL + "/api.php"
	userAgent      = "GW2-MCP-Server/1.0.0"
	requestTimeout = 30 * time.Second
)

// Client handles wiki API requests
type Client struct {
	httpClient *http.Client
	cache      *cache.Manager
	logger     *log.Logger
}

// SearchResult represents a single search result from the wiki
type SearchResult struct {
	Title     string `json:"title"`
	PageID    int    `json:"pageid"`
	Size      int    `json:"size"`
	WordCount int    `json:"wordcount"`
	Snippet   string `json:"snippet"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`
	Extract   string `json:"extract,omitempty"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Query      string         `json:"query"`
	Results    []SearchResult `json:"results"`
	Total      int            `json:"total"`
	SearchedAt time.Time      `json:"searched_at"`
}

// WikiAPIResponse represents the MediaWiki API response structure
type WikiAPIResponse struct {
	BatchComplete string `json:"batchcomplete"`
	Query         struct {
		Search []struct {
			NS        int    `json:"ns"`
			Title     string `json:"title"`
			PageID    int    `json:"pageid"`
			Size      int    `json:"size"`
			WordCount int    `json:"wordcount"`
			Snippet   string `json:"snippet"`
			Timestamp string `json:"timestamp"`
		} `json:"search"`
		SearchInfo struct {
			TotalHits int `json:"totalhits"`
		} `json:"searchinfo"`
	} `json:"query"`
}

// PageContentResponse represents page content API response
type PageContentResponse struct {
	BatchComplete string `json:"batchcomplete"`
	Query         struct {
		Pages map[string]struct {
			PageID  int    `json:"pageid"`
			NS      int    `json:"ns"`
			Title   string `json:"title"`
			Extract string `json:"extract"`
		} `json:"pages"`
	} `json:"query"`
}

// NewClient creates a new wiki client
func NewClient(cacheManager *cache.Manager, logger *log.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: requestTimeout,
		},
		cache:  cacheManager,
		logger: logger,
	}
}

// Search performs a search on the Guild Wars 2 wiki
func (c *Client) Search(ctx context.Context, query string, limit int) (*SearchResponse, error) {
	// Normalize query for caching
	normalizedQuery := strings.ToLower(strings.TrimSpace(query))
	cacheKey := c.cache.GetWikiSearchKey(normalizedQuery)

	// Try cache first
	var searchResponse SearchResponse
	if c.cache.GetJSON(cacheKey, &searchResponse) {
		c.logger.Debug("Wiki search cache hit", "query", query)
		return &searchResponse, nil
	}

	c.logger.Debug("Wiki search cache miss, fetching from API", "query", query)

	// Perform search
	searchResults, err := c.performSearch(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Enhance results with page extracts
	for i := range searchResults {
		extract, err := c.getPageExtract(ctx, searchResults[i].Title)
		if err != nil {
			c.logger.Warn("Failed to get page extract", "title", searchResults[i].Title, "error", err)
		} else {
			searchResults[i].Extract = extract
		}
		searchResults[i].URL = fmt.Sprintf("%s/wiki/%s", wikiBaseURL, url.QueryEscape(searchResults[i].Title))
	}

	// Create response
	searchResponse = SearchResponse{
		Query:      query,
		Results:    searchResults,
		Total:      len(searchResults),
		SearchedAt: time.Now(),
	}

	// Cache the result
	if err := c.cache.SetJSON(cacheKey, searchResponse, cache.WikiDataTTL); err != nil {
		c.logger.Warn("Failed to cache search results", "error", err)
	}

	return &searchResponse, nil
}

// performSearch makes the actual search API call
func (c *Client) performSearch(ctx context.Context, query string, limit int) ([]SearchResult, error) {
	// Build search URL
	params := url.Values{
		"action":   {"query"},
		"format":   {"json"},
		"list":     {"search"},
		"srsearch": {query},
		"srlimit":  {fmt.Sprintf("%d", limit)},
		"srprop":   {"size|wordcount|timestamp|snippet"},
	}

	searchURL := fmt.Sprintf("%s?%s", wikiAPIURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("wiki API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var apiResponse WikiAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Convert to our format
	results := make([]SearchResult, len(apiResponse.Query.Search))
	for i, item := range apiResponse.Query.Search {
		results[i] = SearchResult{
			Title:     item.Title,
			PageID:    item.PageID,
			Size:      item.Size,
			WordCount: item.WordCount,
			Snippet:   c.cleanSnippet(item.Snippet),
			Timestamp: item.Timestamp,
		}
	}

	return results, nil
}

// getPageExtract retrieves a short extract for a wiki page
func (c *Client) getPageExtract(ctx context.Context, title string) (string, error) {
	cacheKey := c.cache.GetWikiPageKey(title)

	// Try cache first
	if extract, found := c.cache.GetString(cacheKey); found {
		return extract, nil
	}

	// Build extract URL
	params := url.Values{
		"action":          {"query"},
		"format":          {"json"},
		"prop":            {"extracts"},
		"titles":          {title},
		"exintro":         {"true"},
		"explaintext":     {"true"},
		"exsectionformat": {"plain"},
		"exchars":         {"500"}, // Limit to 500 characters
	}

	extractURL := fmt.Sprintf("%s?%s", wikiAPIURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", extractURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("extract API request failed with status %d", resp.StatusCode)
	}

	var contentResponse PageContentResponse
	if err := json.NewDecoder(resp.Body).Decode(&contentResponse); err != nil {
		return "", fmt.Errorf("failed to decode extract response: %w", err)
	}

	// Extract the content
	var extract string
	for _, page := range contentResponse.Query.Pages {
		extract = page.Extract
		break // Take the first (and should be only) page
	}

	// Cache the extract
	c.cache.Set(cacheKey, extract, cache.WikiDataTTL)

	return extract, nil
}

// cleanSnippet removes HTML tags and cleans up the snippet text
func (c *Client) cleanSnippet(snippet string) string {
	// Remove HTML tags
	cleaned := strings.ReplaceAll(snippet, "<span class=\"searchmatch\">", "")
	cleaned = strings.ReplaceAll(cleaned, "</span>", "")
	cleaned = strings.ReplaceAll(cleaned, "&quot;", "\"")
	cleaned = strings.ReplaceAll(cleaned, "&amp;", "&")
	cleaned = strings.ReplaceAll(cleaned, "&lt;", "<")
	cleaned = strings.ReplaceAll(cleaned, "&gt;", ">")

	// Clean up whitespace
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.ReplaceAll(cleaned, "\n", " ")
	cleaned = strings.ReplaceAll(cleaned, "\t", " ")

	// Remove multiple spaces
	for strings.Contains(cleaned, "  ") {
		cleaned = strings.ReplaceAll(cleaned, "  ", " ")
	}

	return cleaned
}

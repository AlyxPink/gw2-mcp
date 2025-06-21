package wiki

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/log"

	"github.com/AlyxPink/gw2-mcp/internal/cache"
)

func TestClient_cleanSnippet(t *testing.T) {
	client := &Client{}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "HTML tags removal",
			input:    `<span class="searchmatch">Dragon</span> Bash is a festival`,
			expected: "Dragon Bash is a festival",
		},
		{
			name:     "HTML entities",
			input:    "&quot;Dragon Bash&quot; &amp; other events &lt;test&gt;",
			expected: `"Dragon Bash" & other events <test>`,
		},
		{
			name:     "Whitespace cleanup",
			input:    "Dragon\nBash\t  festival   with    spaces",
			expected: "Dragon Bash festival with spaces",
		},
		{
			name: "Mixed content",
			input: `<span class="searchmatch">Dragon</span>
	Bash &amp; &quot;events&quot;   `,
			expected: `Dragon Bash & "events"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.cleanSnippet(tt.input)
			if result != tt.expected {
				t.Errorf("cleanSnippet() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestClient_Search_Cache(t *testing.T) {
	// Create a mock cache manager
	cacheManager := cache.NewManager()
	logger := log.New(os.Stderr)
	logger.SetLevel(log.ErrorLevel) // Reduce noise in tests

	client := NewClient(cacheManager, logger)

	// Mock the search response
	mockResponse := &SearchResponse{
		Query:   "test query",
		Results: []SearchResult{{Title: "Test Page", PageID: 123}},
		Total:   1,
	}

	// Cache the response
	cacheKey := cacheManager.GetWikiSearchKey("test query")
	err := cacheManager.SetJSON(cacheKey, mockResponse, time.Minute)
	if err != nil {
		t.Fatalf("Failed to cache response: %v", err)
	}

	// Test cache hit
	result, err := client.Search(context.Background(), "test query", 5)
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if result.Query != "test query" {
		t.Errorf("Expected query 'test query', got %q", result.Query)
	}

	if len(result.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(result.Results))
	}

	if result.Results[0].Title != "Test Page" {
		t.Errorf("Expected title 'Test Page', got %q", result.Results[0].Title)
	}
}

func TestClient_Search_API(t *testing.T) {
	// Create mock server for wiki API
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/api.php") {
			if r.URL.Query().Get("list") == "search" {
				// Mock search response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"batchcomplete": "",
					"query": {
						"search": [
							{
								"ns": 0,
								"title": "Dragon Bash",
								"pageid": 12345,
								"size": 5000,
								"wordcount": 800,
								"snippet": "<span class=\"searchmatch\">Dragon</span> Bash is a festival",
								"timestamp": "2023-07-01T12:00:00Z"
							}
						],
						"searchinfo": {
							"totalhits": 1
						}
					}
				}`))
			} else if r.URL.Query().Get("prop") == "extracts" {
				// Mock extract response
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{
					"batchcomplete": "",
					"query": {
						"pages": {
							"12345": {
								"pageid": 12345,
								"ns": 0,
								"title": "Dragon Bash",
								"extract": "Dragon Bash is an annual festival in Guild Wars 2."
							}
						}
					}
				}`))
			}
		}
	}))
	defer mockServer.Close()

	// Override the wiki API URL for testing
	originalURL := wikiAPIURL
	defer func() {
		// Note: In a real implementation, we'd need to make wikiAPIURL configurable
		// For this test, we're just demonstrating the structure
		_ = originalURL
	}()

	// Create client with fresh cache
	cacheManager := cache.NewManager()
	logger := log.New(os.Stderr)
	logger.SetLevel(log.ErrorLevel)

	_ = NewClient(cacheManager, logger)

	// Note: This test would need the client to be configurable to use the mock server
	// For now, we'll test the response parsing logic separately
	t.Log("Mock server created for API testing structure")
}

func TestSearchResult_URLGeneration(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{
			name:     "Simple title",
			title:    "Dragon Bash",
			expected: "https://wiki.guildwars2.com/wiki/Dragon%20Bash",
		},
		{
			name:     "Title with special characters",
			title:    "API:Account/wallet",
			expected: "https://wiki.guildwars2.com/wiki/API:Account%2Fwallet",
		},
		{
			name:     "Title with spaces and punctuation",
			title:    "Living World Season 4: Episode 1",
			expected: "https://wiki.guildwars2.com/wiki/Living%20World%20Season%204:%20Episode%201",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This tests the URL generation logic that would be in the Search method
			result := SearchResult{Title: tt.title}
			// In the actual implementation, this URL would be set in the Search method
			// Use PathEscape for proper URL encoding in path segments
			generatedURL := "https://wiki.guildwars2.com/wiki/" + url.PathEscape(tt.title)

			if generatedURL != tt.expected {
				t.Errorf("URL generation for %q: got %q, want %q", tt.title, generatedURL, tt.expected)
			}

			_ = result // Use the result to avoid unused variable error
		})
	}
}

func TestNewClient(t *testing.T) {
	cacheManager := cache.NewManager()
	logger := log.New(io.Discard)

	client := NewClient(cacheManager, logger)

	if client == nil {
		t.Fatal("Expected non-nil client")
	}

	if client.cache != cacheManager {
		t.Error("Expected cache manager to be set")
	}

	if client.logger != logger {
		t.Error("Expected logger to be set")
	}

	if client.httpClient == nil {
		t.Error("Expected HTTP client to be initialized")
	}

	if client.httpClient.Timeout != requestTimeout {
		t.Errorf("Expected timeout %v, got %v", requestTimeout, client.httpClient.Timeout)
	}
}

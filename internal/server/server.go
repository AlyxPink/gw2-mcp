package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"

	"github.com/AlyxPink/gw2-mcp/internal/cache"
	"github.com/AlyxPink/gw2-mcp/internal/gw2api"
	"github.com/AlyxPink/gw2-mcp/internal/wiki"

	"github.com/charmbracelet/log"
)

// MCPServer wraps the MCP server with GW2-specific functionality
type MCPServer struct {
	mcp    *mcpserver.MCPServer
	logger *log.Logger
	cache  *cache.Manager
	gw2API *gw2api.Client
	wiki   *wiki.Client
}

// NewMCPServer creates a new GW2 MCP server instance
func NewMCPServer(logger *log.Logger) (*MCPServer, error) {
	// Create cache manager
	cacheManager := cache.NewManager()

	// Create GW2 API client
	gw2Client := gw2api.NewClient(cacheManager, logger)

	// Create wiki client
	wikiClient := wiki.NewClient(cacheManager, logger)

	// Create MCP server
	mcpServer := mcpserver.NewMCPServer(
		"GW2 MCP Server",
		"1.0.0",
		mcpserver.WithToolCapabilities(true),
		mcpserver.WithResourceCapabilities(true, true),
		mcpserver.WithRecovery(),
	)

	gw2MCP := &MCPServer{
		mcp:    mcpServer,
		logger: logger,
		cache:  cacheManager,
		gw2API: gw2Client,
		wiki:   wikiClient,
	}

	// Register tools
	if err := gw2MCP.registerTools(); err != nil {
		return nil, fmt.Errorf("failed to register tools: %w", err)
	}

	// Register resources
	if err := gw2MCP.registerResources(); err != nil {
		return nil, fmt.Errorf("failed to register resources: %w", err)
	}

	return gw2MCP, nil
}

// Start starts the MCP server
func (s *MCPServer) Start(ctx context.Context) error {
	s.logger.Info("Starting MCP server on stdio")
	return mcpserver.ServeStdio(s.mcp)
}

// registerTools registers all available tools
func (s *MCPServer) registerTools() error {
	// Wiki search tool
	wikiSearchTool := mcp.NewTool(
		"wiki_search",
		mcp.WithDescription("Search Guild Wars 2 wiki for information about game content"),
		mcp.WithString(
			"query",
			mcp.Required(),
			mcp.Description("Search query for wiki content (e.g., 'Dragon Bash', 'currencies', 'wallet')"),
		),
		mcp.WithNumber(
			"limit",
			mcp.Description("Maximum number of results to return (default: 5)"),
		),
	)

	s.mcp.AddTool(wikiSearchTool, s.handleWikiSearch)

	// Wallet info tool
	walletTool := mcp.NewTool(
		"get_wallet",
		mcp.WithDescription("Get user's wallet information including all currencies"),
		mcp.WithString(
			"api_key",
			mcp.Required(),
			mcp.Description("Guild Wars 2 API key with account scope"),
		),
	)

	s.mcp.AddTool(walletTool, s.handleGetWallet)

	// Currency info tool
	currencyTool := mcp.NewTool(
		"get_currencies",
		mcp.WithDescription("Get information about Guild Wars 2 currencies"),
		mcp.WithArray(
			"ids",
			mcp.Description("Specific currency IDs to fetch (optional, returns all if not specified)"),
		),
	)

	s.mcp.AddTool(currencyTool, s.handleGetCurrencies)

	return nil
}

// registerResources registers all available resources
func (s *MCPServer) registerResources() error {
	// Currency list resource
	currencyListResource := mcp.NewResource(
		"gw2://currencies",
		"Guild Wars 2 Currencies",
		mcp.WithResourceDescription("Complete list of all Guild Wars 2 currencies with metadata"),
		mcp.WithMIMEType("application/json"),
	)

	s.mcp.AddResource(currencyListResource, s.handleCurrencyListResource)

	return nil
}

package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// handleWikiSearch handles wiki search requests
func (s *MCPServer) handleWikiSearch(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	query, err := request.RequireString("query")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid query parameter: %v", err)), nil
	}

	// Get limit parameter (optional)
	const defaultLimit = 5
	limit := request.GetInt("limit", defaultLimit)

	s.logger.Debug("Wiki search request", "query", query, "limit", limit)

	// Perform wiki search
	results, err := s.wiki.Search(ctx, query, limit)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Wiki search failed: %v", err)), nil
	}

	// Format results as JSON
	resultJSON, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(resultJSON)), nil
}

// handleGetWallet handles wallet information requests
func (s *MCPServer) handleGetWallet(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	apiKey, err := request.RequireString("api_key")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid api_key parameter: %v", err)), nil
	}

	s.logger.Debug("Wallet request", "api_key_length", len(apiKey))

	// Get wallet information
	wallet, err := s.gw2API.GetWallet(ctx, apiKey)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get wallet: %v", err)), nil
	}

	// Format wallet as JSON
	walletJSON, err := json.MarshalIndent(wallet, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format wallet: %v", err)), nil
	}

	return mcp.NewToolResultText(string(walletJSON)), nil
}

// handleGetCurrencies handles currency information requests
func (s *MCPServer) handleGetCurrencies(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Parse optional currency IDs
	currencyIDs := request.GetIntSlice("ids", nil)

	s.logger.Debug("Currency request", "currency_ids", currencyIDs)

	// Get currency information
	currencies, err := s.gw2API.GetCurrencies(ctx, currencyIDs)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to get currencies: %v", err)), nil
	}

	// Format currencies as JSON
	currenciesJSON, err := json.MarshalIndent(currencies, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to format currencies: %v", err)), nil
	}

	return mcp.NewToolResultText(string(currenciesJSON)), nil
}

// handleCurrencyListResource handles the currency list resource
func (s *MCPServer) handleCurrencyListResource(ctx context.Context,
	_ mcp.ReadResourceRequest,
) ([]mcp.ResourceContents, error) {
	s.logger.Debug("Currency list resource request")

	// Get all currencies
	currencies, err := s.gw2API.GetCurrencies(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get currencies: %w", err)
	}

	// Format currencies as JSON
	currenciesJSON, err := json.MarshalIndent(currencies, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format currencies: %w", err)
	}

	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "gw2://currencies",
			MIMEType: "application/json",
			Text:     string(currenciesJSON),
		},
	}, nil
}

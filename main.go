// Package main provides the GW2 MCP (Model Context Protocol) server implementation.
// This server provides access to Guild Wars 2 API data and wiki information through MCP.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/charmbracelet/log"

	"github.com/AlyxPink/gw2-mcp/internal/server"
)

func main() {
	// Setup logger
	logger := log.NewWithOptions(os.Stderr, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		Level:           log.DebugLevel,
	})

	// Create context that cancels on interrupt
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		logger.Info("Shutting down gracefully...")
		cancel()
	}()

	// Create and start the MCP server
	mcpServer, err := server.NewMCPServer(logger)
	if err != nil {
		logger.Fatal("Failed to create MCP server", "error", err)
	}

	logger.Info("Starting GW2 MCP Server")
	if err := mcpServer.Start(ctx); err != nil {
		logger.Fatal("Server failed", "error", err)
	}

	fmt.Println("Server stopped")
}

# GW2 Model Context Provider Rules

## Technical Requirements

- Go version >= 1.24
- If needed, latest github.com/charmbracelet libraries (including alpha/beta releases):
  - github.com/charmbracelet/bubbletea
  - github.com/charmbracelet/lipgloss
  - github.com/charmbracelet/log
  - etc
- https://github.com/mark3labs/mcp-go or any other very popular framework to build a MCP server using Golang

## Project Goals

Create a Model Context Provider (MCP) server for Guild Wars 2 that:

1. Provides context and data to Large Language Models (LLMs)
2. Implements core functionality:
   - Wiki page search and content retrieval
   - GW2 API integration for user data
   - Wallet and currency information
   - Extensible architecture for future features

## Design Philosophy

- Focus on reliability and data accuracy
- Build modular and extensible components
- Ensure clear API documentation
- Implement robust error handling
- Keep the codebase simple and maintainable

## Code Standards

1. Clean Architecture principles
2. Comprehensive error handling
3. Unit tests for core functionality
4. Documentation for all public APIs
5. Consistent code formatting using `gofumpt` and `golangci-lint`
6. Meaningful commit messages following conventional commits

## Performance Guidelines

- Efficient API calls to GW2 servers and Wiki
- Smart caching for frequently accessed data
- Rate limiting implementation
- Resource-conscious operation
- Quick response times for LLM queries

## API Design

- RESTful endpoints for LLM interaction
- Clear request/response structures
- Proper validation and sanitization
- Comprehensive error responses
- API versioning support

## Development Workflow

1. Feature branches for new functionality
2. Pull request reviews
3. CI/CD pipeline checks
4. Regular dependency updates
5. Version tagging for releases

These rules ensure we create a reliable and efficient Model Context Provider that can serve as a bridge between LLMs and Guild Wars 2 data sources.

# GW2 MCP Server

A Model Context Provider (MCP) server for Guild Wars 2 that bridges Large Language Models (LLMs) with Guild Wars 2 data sources.

## Features

- **Wiki Search**: Search and retrieve content from the Guild Wars 2 wiki
- **Wallet Information**: Access user wallet and currency data via GW2 API
- **Smart Caching**: Efficient caching with appropriate TTL for static and dynamic data
- **Rate Limiting**: Respectful API usage with built-in rate limiting
- **Extensible Architecture**: Modular design for easy feature additions

## Requirements

- Go 1.24 or higher
- Guild Wars 2 API key (for wallet functionality)

## Installation

1. Clone the repository:
```bash
git clone https://github.com/AlyxPink/gw2-mcp.git
cd gw2-mcp
```

2. Install dependencies:
```bash
go mod tidy
```

3. Build the server:
```bash
go build -o gw2-mcp ./cmd/server
```

## Usage

### Running the Server

The MCP server communicates via stdio (standard input/output):

```bash
./gw2-mcp
```

### MCP Tools

The server provides the following tools for LLM interaction:

#### 1. Wiki Search (`wiki_search`)

Search the Guild Wars 2 wiki for information.

**Parameters:**
- `query` (required): Search query string
- `limit` (optional): Maximum number of results (default: 5)

**Example:**
```json
{
  "tool": "wiki_search",
  "arguments": {
    "query": "Dragon Bash",
    "limit": 3
  }
}
```

#### 2. Get Wallet (`get_wallet`)

Retrieve user's wallet information including all currencies.

**Parameters:**
- `api_key` (required): Guild Wars 2 API key with account scope

**Example:**
```json
{
  "tool": "get_wallet",
  "arguments": {
    "api_key": "YOUR_GW2_API_KEY"
  }
}
```

#### 3. Get Currencies (`get_currencies`)

Get information about Guild Wars 2 currencies.

**Parameters:**
- `ids` (optional): Array of specific currency IDs to fetch

**Example:**
```json
{
  "tool": "get_currencies",
  "arguments": {
    "ids": [1, 2, 3]
  }
}
```

### MCP Resources

The server provides the following resources:

#### Currency List (`gw2://currencies`)

Complete list of all Guild Wars 2 currencies with metadata.

## API Key Setup

To use wallet functionality, you need a Guild Wars 2 API key:

1. Visit [Guild Wars 2 API Key Management](https://account.arena.net/applications)
2. Create a new API key with the following permissions:
   - `account` - Required for wallet access
   - `wallet` - Required for currency information
3. Copy the generated API key

**Security Note:** API keys are hashed before caching for security. Never share your API key.

## Caching Strategy

The server implements intelligent caching:

- **Static Data** (currencies, wiki content): Cached for 24 hours to 1 year
- **Dynamic Data** (wallet balances): Cached for 5 minutes
- **Search Results**: Cached for 24 hours

## Architecture

The project follows Clean Architecture principles:

```
internal/
├── server/          # MCP server implementation
├── cache/           # Caching layer
├── gw2api/          # GW2 API client
└── wiki/            # Wiki API client
```

## Development

### Code Standards

- Format code with `gofumpt`
- Lint with `golangci-lint`
- Write unit tests for core functionality
- Follow conventional commit messages

### Running Tests

```bash
go test ./...
```

### Linting

```bash
golangci-lint run
```

### Formatting

```bash
gofumpt -w .
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Run linting and formatting
6. Submit a pull request

## License

MIT License - see LICENSE file for details.

## Acknowledgments

- [Guild Wars 2 API](https://wiki.guildwars2.com/wiki/API:Main) for providing comprehensive game data
- [Guild Wars 2 Wiki](https://wiki.guildwars2.com/) for extensive game documentation
- [MCP Go](https://github.com/mark3labs/mcp-go) for the MCP implementation framework
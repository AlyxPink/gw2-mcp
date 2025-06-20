# MCP Server for Guild Wars 2: Prompt for IDE Code Editor

**Objective**:
Build a **Model Context Provider (MCP) server** using Go that bridges Large Language Models (LLMs) with Guild Wars 2 data sources. The server must provide accurate, cached context for LLMs through two core tools:

1. A **wiki search API** to retrieve and contextualize GW2 content from [https://wiki.guildwars2.com/](https://wiki.guildwars2.com/)
2. A **user wallet/currency API** to fetch user-specific data via the Guild Wars 2 v2 API (e.g., https://api.guildwars2.com/v2/account/wallet)

---

### Technical Requirements

- Use **Go ≥1.24** with modern tooling: `gofumpt`, `golangci-lint`, and dependencies from [https://github.com/mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) or any other **popular Go framework** (e.g., Echo, Gin, or FastHTTP).
- Integrate with the following libraries for terminal UI/UX:
  - `github.com/charmbracelet/bubbletea` (for interactive CLI tools if needed)
  - `github.com/charmbracelet/lipgloss` (for styling)
  - `github.com/charmbracelet/log` (for logging)

---

### Core Functionalities

1. **Wiki Search API**:

   - Accept search terms and return relevant wiki content from [https://wiki.guildwars2.com/wiki/Main_Page](https://wiki.guildwars2.com/wiki/Main_Page).
   - Cache **static data** (e.g., currency names) for extended periods (ideally forever) to avoid redundant API calls.
   - If a query returns no match, fetch and cache the result for future use.

2. **Wallet & Currency API**:

   - Use user-provided API tokens to access [https://api.guildwars2.com/v2/account/wallet](https://api.guildwars2.com/v2/account/wallet) and [https://api.guildwars2.com/v2/currencies](https://api.guildwars2.com/v2/currencies).
   - Cache **currency metadata** (names, IDs) for long-term use.
   - Handle rate limiting and error recovery for API requests.

3. **Caching Strategy**:
   - Implement a robust caching layer (e.g., Redis or in-memory cache) to reduce load on external services.
   - For static data (e.g., currency names), prioritize **long-term caching** (e.g., 1 year).
   - Use **TTL (Time-to-Live)** for dynamic data (e.g., wallet balances) based on API response headers.

---

### Code Standards & Architecture

- Follow **Clean Architecture** principles: separate domain logic, application services, and infrastructure layers.
- Ensure **comprehensive error handling**: include retry logic, fallbacks, and clear error messages for users.
- Write **unit tests** for all core functionality (e.g., caching, API calls).
- Document all public APIs with **Swagger/OpenAPI spec** and versioning support (e.g., `/api/v1/wallet`).
- Format code using `gofumpt` and lint with `golangci-lint`.

---

### Performance Guidelines

- Optimize API calls to minimize latency: use **batch requests**, **caching**, and **rate limiting**.
- Implement **resource-conscious design**: avoid memory leaks, manage goroutines efficiently, and prioritize low-latency responses for LLMs.
- Add **metrics tracking** (e.g., request counts, cache hit rates) for monitoring and optimization.

---

### API Design Principles

1. **RESTful Endpoints**:

   - Example: `/api/v1/wiki/search?query=Dragon`
   - `/api/v1/wallet` (GET with token authentication)
   - `/api/v1/currencies` (GET for metadata, POST to fetch user-specific data)

2. **Versioning**: Support API versioning (e.g., `v1`, `v2`) to ensure backward compatibility.

3. **Validation & Sanitization**: Validate all inputs (e.g., tokens, search terms) and sanitize outputs.

4. **Error Responses**: Return clear HTTP status codes and error messages (e.g., 401 for invalid tokens).

---

### Extensibility & Maintainability

- Design the server to support future features (e.g., guild data, item stats, or LLM query processing).
- Use modular components (e.g., separate packages for caching, API clients, and domain logic).
- Follow **conventional commit messages** and maintain a clear Git history.

---

### Development Workflow

1. Create **feature branches** for new functionality.
2. Include **CI/CD checks** for code quality (e.g., `gofumpt`, `golangci-lint`) and test coverage.
3. Tag releases with semantic versioning (e.g., `v0.1.0`).

---

### Final Notes

- Ensure the server is **secure**: validate tokens, sanitize inputs, and avoid exposing sensitive data.
- Prioritize **data accuracy** for LLM context: cache static content but fetch dynamic user data via API.
- Document all public APIs in detail for LLM integration (e.g., how to query currency names or wallet balances).

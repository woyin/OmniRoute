# OmniRoute (Go Rewrite)

Unified AI proxy/router — route any LLM through one endpoint.  
Go rewrite of the original TypeScript/Next.js OmniRoute.

## Status

**Early implementation** — core routing, 6 priority providers, SQLite persistence, and SSE streaming are functional.

### Implemented Providers

| Provider | ID | Auth Type | Status |
|----------|-----|-----------|--------|
| OpenCode Free (zen) | `opencode` | NoAuth | ✅ |
| OpenCode Go | `opencode-go` | API Key | ✅ |
| Ollama Cloud | `ollama-cloud` | API Key | ✅ |
| Codex (OpenAI) | `codex` | OAuth | ✅ |
| Command Code | `command-code` | API Key | ✅ |
| OpenAI Compatible | `openai-compatible-*` | API Key | ✅ |
| Anthropic Compatible | `anthropic-compatible-*` | API Key | ✅ |

### Implemented Features

- **HTTP Server**: Chi router with CORS, compression, recovery middleware
- **Chat Completions API**: `/api/v1/chat/completions` (streaming + non-streaming)
- **Responses API**: `/api/v1/responses`
- **Models API**: `/api/v1/models`
- **Provider Registry**: In-memory registry with model catalogs
- **Executor Framework**: Provider-specific executors with retry logic
- **Request Translation**: OpenAI ↔ Claude ↔ Gemini format translation
- **SSE Streaming**: Full streaming support with heartbeats
- **Combo Routing**: 5 strategies (priority, weighted, round-robin, random, fill-first)
- **SQLite Persistence**: WAL mode, migrations, provider connections, combos, API keys
- **API Key Auth**: Optional and required modes
- **A2A Agent Card**: `/.well-known/agent.json`
- **Management API**: Provider connections, combos, API keys CRUD

### Not Yet Implemented (from original)

- MCP Server (94 tools, stdio/SSE/Streamable HTTP)
- A2A full protocol (JSON-RPC, task lifecycle)
- Prompt compression pipeline
- Memory system
- Skills system
- Guardrails framework
- Cloud agents
- Webhooks
- Evals framework
- Tunnels (Cloudflare, ngrok)
- All 237 providers (only 7 implemented so far)
- Electron desktop app

## Quick Start

```bash
# Build
cd go-rewrite
go build -o omniroute ./cmd/omniroute

# Run with defaults (port 3456)
./omniroute

# Custom port
./omniroute -port 8080

# Custom data directory
./omniroute -data-dir /path/to/data

# With environment variables
PORT=8080 REQUIRE_API_KEY=true ./omniroute
```

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Health check |
| POST | `/api/v1/chat/completions` | Chat completions (OpenAI format) |
| POST | `/api/v1/responses` | Responses API |
| GET | `/api/v1/models` | List available models |
| GET | `/api/providers` | List provider connections |
| POST | `/api/providers` | Create provider connection |
| GET | `/api/combos` | List combos |
| POST | `/api/combos` | Create combo |
| GET | `/api/api-keys` | List API keys |
| POST | `/api/api-keys` | Create API key |
| GET | `/.well-known/agent.json` | A2A agent card |

## Architecture

```
cmd/omniroute/          # Entry point
internal/
  config/               # Configuration (env vars, defaults)
  db/                   # SQLite persistence (migrations, domain modules)
  handler/              # HTTP route handlers (chat, health, models, combos)
  middleware/           # CORS, recovery, logging
  auth/                 # API key authentication
  provider/
    executor/           # Provider-specific executors (opencode, codex, command-code, default)
    registry/           # Provider registry (builtin + custom)
    translator/         # Format translation (OpenAI ↔ Claude ↔ Gemini)
  routing/              # Combo routing engine (5 strategies)
  sse/                  # SSE streaming utilities
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | 3456 | Server port |
| `DATA_DIR` | `~/.omniroute` | Data directory |
| `REQUIRE_API_KEY` | false | Require API key for all requests |
| `FETCH_TIMEOUT_MS` | 120000 | Upstream fetch timeout |
| `STREAM_IDLE_TIMEOUT_MS` | 60000 | SSE stream idle timeout |
| `SSE_HEARTBEAT_INTERVAL_MS` | 15000 | SSE heartbeat interval |
| `CODEX_OAUTH_CLIENT_ID` | | Codex OAuth client ID |
| `CODEX_OAUTH_CLIENT_SECRET` | | Codex OAuth client secret |
| `OPENCODE_SYNTHESIZE_CLI_HEADERS` | false | Synthesize OpenCode CLI identity headers |

## Testing

```bash
go test ./...
```

## Comparison with TypeScript Version

| Aspect | TypeScript (original) | Go (rewrite) |
|--------|----------------------|--------------|
| Runtime | Node.js 22+ | Go 1.24 |
| Framework | Next.js 16 (App Router) | Chi router |
| Database | better-sqlite3 | go-sqlite3 |
| Providers | 237 | 7 (priority set) |
| Streaming | SSE via open-sse | Native SSE |
| Schema validation | Zod v4 | JSON decode + type assertions |
| Build output | `.build/next/` + `dist/` | Single binary (~14MB) |
| Startup time | ~3-5s | <100ms |

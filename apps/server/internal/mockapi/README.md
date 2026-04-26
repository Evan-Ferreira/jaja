# Mock Anthropic API Server

A standalone mock server for testing JAJA without consuming Anthropic API quota. Simulates the Anthropic API with embedded JSON fixtures.

## Quick Start

### Via Docker

```bash
docker compose up mock-api
```

Runs on http://localhost:4010 (configured via `MOCKAPI_PORT=4010` in docker-compose.yml).

### Manual with Air (hot reload)

```bash
cd apps/server
air -c .air.mockapi.toml
```

Runs on the port specified by `MOCKAPI_PORT` env var (default 4010).

## Configuration

Set these environment variables in `apps/server/.env`:

| Variable | Default | Purpose |
| --- | --- | --- |
| `MOCKAPI_PORT` | `4010` | Port the mock server listens on |
| `ANTHROPIC_API_URL` | `https://api.anthropic.com` | Set this on the **main server** to point at mock-api (e.g., `http://mock-api:4010` in Docker, `http://localhost:4010` locally) |

**Point the main server at mock-api:**

In `apps/server/.env`:

```bash
ANTHROPIC_API_URL=http://localhost:4010      # For manual testing
ANTHROPIC_API_URL=http://mock-api:4010       # For Docker
```

## Available Endpoints

All routes are prefixed with `/v1` (matching Anthropic API versioning).

| Method | Path | Status | Handler |
| --- | --- | --- | --- |
| `POST` | `/v1/messages` | ✅ Working | Returns embedded JSON fixture (`analyze_basic.json`) |
| `GET` | `/v1/files/:id` | ⚠️ Stub | Returns 404 (file metadata endpoint) |
| `GET` | `/v1/files/:id/content` | ⚠️ Stub | Returns 404 (file content endpoint) |

## Fixtures

Fixtures are JSON responses embedded into the binary at build time using Go's `//go:embed` directive. They live in `handlers/anthropic/testdata/`.

### Current Fixtures

- **`analyze_basic.json`** — Response for `/v1/messages`, simulates Claude analyzing an assignment

### Fixture Format

Fixtures are valid Anthropic API response JSONs. Example structure:

```json
{
  "id": "msg_01TestFixture",
  "type": "message",
  "role": "assistant",
  "model": "claude-haiku-4-5-20251001",
  "content": [
    {
      "type": "text",
      "text": "Analysis of the assignment..."
    }
  ],
  "stop_reason": "end_turn",
  "stop_sequence": null,
  "usage": {
    "input_tokens": 150,
    "output_tokens": 200
  }
}
```

## How to Add a Fixture

1. Create a new JSON file in `handlers/anthropic/testdata/` (e.g., `my_fixture.json`)

2. Update `handlers/anthropic/provider.go` to embed and use it:

```go
import _ "embed"

// TODO: make this configurable based on request content
//go:embed testdata/my_fixture.json
var myFixture []byte

func HandleMessages(c *gin.Context) {
    response := map[string]any{}
    err := json.Unmarshal(myFixture, &response)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, response)
}
```

3. Rebuild the binary (Air auto-rebuilds when you save Go files if using manual mode)

4. Test:
```bash
curl -X POST http://localhost:4010/v1/messages \
  -H "Content-Type: application/json" \
  -d '{"model": "claude-3-5-sonnet-20241022", "max_tokens": 1000, "messages": []}'
```

## How to Add an Endpoint

1. Create a handler function in `handlers/anthropic/provider.go`:

```go
func HandleNewEndpoint(c *gin.Context) {
    // Return a fixture, custom response, or error
    c.JSON(200, gin.H{
        "key": "value",
        "status": "ok",
    })
}
```

2. Register the route in `routes/anthropic.go`:

```go
func RegisterAnthropicRoutes(rg *gin.RouterGroup) {
    routes := rg.Group("/v1")
    {
        routes.POST("/messages", anthropic.HandleMessages)
        routes.GET("/new-endpoint", anthropic.HandleNewEndpoint)  // Add this line
    }
}
```

3. Rebuild and test:

```bash
curl http://localhost:4010/v1/new-endpoint
```

## Testing with the Main Server

Once the mock server is running, the main JAJA server will use it for all Anthropic API calls. This allows you to:

- **Test agent workflows** without API quota costs
- **Test with deterministic responses** (same fixture every time)
- **Iterate quickly** on agent logic without network latency
- **Develop offline** if Anthropic API is unavailable

Example: Run the agent with mock responses:

```bash
# Terminal 1: Start mock-api
cd apps/server && air -c .air.mockapi.toml

# Terminal 2: Start main server (will use mock-api)
cd apps/server && go run cmd/main.go

# Terminal 3: Run the agent
curl -X POST http://localhost:8080/dev/run-agent \
  -H "Content-Type: application/json" \
  -d '{"session_id": "test", "user_id": "user-1", "assignment_key": "assignment-1"}'
```

## Limitations & TODOs

- File endpoints (`/v1/files/:id`, `/v1/files/:id/content`) currently return 404 stubs
- Fixtures are static — no dynamic request-based response selection yet
- No support for streaming responses or tool use yet
- Consider making fixture selection conditional on request content (e.g., based on prompt parameters)

## Architecture

```
mockapi/
├── cmd/mockapi/main.go          # Entry point: creates Gin router, registers routes
├── internal/mockapi/
│   ├── routes/
│   │   └── anthropic.go          # Route registration: POST /v1/messages, GET /v1/files/:id
│   └── handlers/
│       └── anthropic/
│           ├── provider.go       # Handler functions: HandleMessages, HandleFileMetadata, HandleFileContent
│           └── testdata/         # Embedded JSON fixtures
│               └── analyze_basic.json
└── .air.mockapi.toml            # Air config for hot reload
```

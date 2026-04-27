# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

JAJA (Just Automate Junk Assignments) — a web app for saving D2L (Desire2Learn) cookies and local storage to a database. Monorepo with a Next.js frontend and Go backend, backed by PostgreSQL, MinIO (S3-compatible object storage), Redis (job queue), and Claude AI (Anthropic) for assignment completion.

## Development Commands

### Full stack (Docker)

```bash
# Terminal 1: Start cloudflared tunnel on port 9000
cloudflared tunnel --url http://localhost:9000

# Terminal 2: Start Docker services (db, redis, minio, server)
docker compose up

# Terminal 3: Start the frontend
cd apps/client && bun dev
```

**Service URLs:**
- Frontend: http://localhost:3000 (from `bun dev` in Terminal 3)
- Server: http://localhost:4000 (via Docker) or http://localhost:8080 (manual)
- Mock API: http://localhost:4010 (mockapi server, mocks Anthropic API responses)
- PostgreSQL: localhost:5432
- Redis: localhost:6379
- MinIO API: http://localhost:9000
- MinIO Console: http://localhost:9001

**Important:** The client must be started separately with `bun dev` after `docker compose up`. Docker starts the backend services (db, redis, minio, mock-api, server).

Note: To use the mock API instead of the real Anthropic API, set `ANTHROPIC_API_URL=http://mock-api:4010` in `apps/server/.env` when running via Docker, or `ANTHROPIC_API_URL=http://localhost:4010` when running locally.

Note: The cloudflared tunnel URL (from Terminal 1) should be set as `MINIO_PUBLIC_URL` in `apps/server/.env` for presigned URLs that Claude AI can access.

### Frontend only (`apps/client/`)

```bash
bun install
bunx skills add vercel-labs/next-browser  # AI agent debugging tools
bun dev                    # Next.js dev server on :3000
bun build                  # Production build
bunx eslint .              # Lint
```

After `bunx skills add vercel-labs/next-browser`, Playwright will prompt to install Chromium. Approve it. This installs `@vercel/next-browser` CLI which gives Claude Code access to React DevTools, component trees, props inspection, and Next.js dev overlay data.

### Backend only (`apps/server/`)

The server runs via Air (hot reload) inside Docker on port 4000. To run manually:

```bash
go mod download
go run cmd/main.go
```

This starts the server on port specified by `PORT` env var (default 8080). The Dockerfile exposes port 4000 and uses Air for live reloading. When running manually, ensure `REDIS_URL` is set in `apps/server/.env` (e.g., `localhost:6379`).

### Developer Tools

For working with background jobs and the job queue, install the asynq CLI:

```bash
go install github.com/hibiken/asynq/tools/asynq@latest
```

This provides the `asynq` command-line tool for inspecting, managing, and debugging Redis queues.

### Agent runner

The JAJA agent is powered by the Google Agent Development Kit (ADK) and orchestrated by `agent/runner/runner.go`. To invoke the agent programmatically or test it:

```bash
# Via HTTP endpoint (preferred)
curl -X POST http://localhost:8080/dev/run-agent \
  -H "Content-Type: application/json" \
  -d '{"session_id":"test-session","prompt":"Complete this assignment...","user_id":"..."}'

# Manual Go testing
go run cmd/main.go  # Server includes agent initialization on startup
```

The agent accepts a prompt (assignment description), processes it with tools (docx generation), and returns the completion result. See `/dev/run-agent` endpoint for usage.

### Mock API server (`apps/server/cmd/mockapi/`)

A standalone Anthropic API mock server for testing without consuming API quota. Runs on port 4010 (in Docker) or configurable via `MOCKAPI_PORT` env var.

**Run mock API server:**

```bash
# Via Docker (auto-started with docker compose up)
docker compose up mock-api

# Manual (with Air hot reload)
cd apps/server
air -c .air.mockapi.toml
```

**Add fixtures:**

Fixtures are embedded JSON responses stored in `internal/mockapi/handlers/anthropic/testdata/`. The `analyze_basic.json` fixture is embedded into the binary via `//go:embed` in `provider.go`:

1. Create a new fixture file in `internal/mockapi/handlers/anthropic/testdata/` (e.g., `my_fixture.json`)
2. Update `provider.go` to embed it and return it from the appropriate handler
3. Example for `/v1/messages` endpoint: modify `HandleMessages()` to select fixtures based on request parameters

**Add new endpoints:**

1. Create handler function in `internal/mockapi/handlers/anthropic/provider.go` (or new handler file)
2. Register route in `internal/mockapi/routes/anthropic.go` by adding to `RegisterAnthropicRoutes()`
3. Example structure:
   ```go
   // In provider.go
   func HandleNewEndpoint(c *gin.Context) {
       // Return fixture or custom response
       c.JSON(200, gin.H{"key": "value"})
   }
   
   // In routes/anthropic.go
   routes.GET("/new-endpoint", anthropic.HandleNewEndpoint)
   ```

### Database migrations (Goose)

Migrations live in `apps/server/migrations/` using Goose SQL format. GORM auto-migrates models on startup, but schema changes should also have corresponding Goose migration files. Environment variables for Goose are in the root `.env`.

## Architecture

```
apps/
├── client/               # Next.js 16 (App Router) + React 19 + Tailwind 4 + shadcn/ui
│   ├── app/              # Pages and app-level components
│   │   ├── dev/files/    # Dev utilities (e.g., assignment file upload)
│   │   └── components/   # Page-specific components (e.g., cookie-form.tsx)
│   ├── components/ui/    # shadcn UI primitives (button, checkbox, field, input, etc.)
│   ├── lib/utils.ts      # cn() helper (clsx + tailwind-merge)
│   └── utils/string.ts   # parseStringToJSON() for tab-separated cookie/storage data
└── server/               # Go 1.25 + Gin + GORM + Anthropic SDK + asynq
    ├── cmd/main.go       # Entry point: loads env, connects DB + Redis + S3 + Agent + Workers, sets up router
    ├── cmd/mockapi/      # Mock Anthropic API server (port 4010)
    │   └── main.go       # Entry point: Gin router with mock handlers
    ├── internal/mockapi/ # Mock API implementation
    │   ├── handlers/
    │   │   └── anthropic/
    │   │       ├── provider.go # Handlers: HandleMessages, HandleFileMetadata, HandleFileContent
    │   │       └── testdata/   # Embedded JSON fixtures (analyze_basic.json)
    │   └── routes/
    │       └── anthropic.go # RegisterAnthropicRoutes - routes `/v1/messages`, `/v1/files/:id`, etc.
    ├── agent/            # Claude AI agent via Google ADK + Anthropic SDK
    │   ├── config.go     # ConnectAgent() initializes global AgentRunner
    │   ├── agents/       # Agent definitions
    │   │   └── orchestrator.go # CreateJAJAAgent() - named agent with docx tool
    │   ├── runner/       # Agent execution
    │   │   └── runner.go # AgentRunner.Run(jobID, userID, assignmentKey) - orchestrates agent, uploads result to S3
    │   ├── tools/        # ADK tool implementations
    │   │   └── docx.go   # create_docx tool - generates .docx Word documents using unioffice
    │   ├── models/       # Model interfaces
    │   │   └── anthropic.go # AnthropicModel — unified Anthropic client: ADK LLM interface + direct Beta Messages API (tools, skills, file download, document blocks)
    │   └── skills/       # Agent utilities and prompts
    ├── internal/
    │   ├── config.go     # Global configuration vars (DBClient, RedisClient, S3BasicsBucket, AgentRunner)
    │   ├── database/     # Database configuration and connection
    │   │   └── config.go # ConnectDB() - initializes PostgreSQL via GORM
    │   ├── queue/        # Redis and job queue configuration
    │   │   └── config.go # ConnectRedis() - initializes Redis for asynq
    │   ├── storage/      # S3/MinIO configuration and helpers
    │   │   ├── config.go # ConnectObjectStorage() - initializes AWS SDK v2 + MinIO
    │   │   └── s3.go     # BucketBasics S3 operations (CRUD, multipart, presigned URLs)
    │   ├── handlers/
    │   │   ├── d2l/      # D2L handlers: SaveCredentials, GetCoursesAndAssignments, SyncCoursesAndAssignments
    │   │   └── dev/      # Dev handlers: SaveAssignmentFiles, GeneratePresignedURL, RunAgent
    │   ├── models/       # GORM models: User, Org, D2LCookieSession, D2LLocalStorageSession, Job
    │   ├── jobs/         # Job queue types and handlers
    │   │   ├── types.go  # JobTypeDocx constant
    │   │   └── handlers/ # Task processors (HandleDocx, HandleUnknown)
    │   ├── workers/      # Database-backed job polling and dispatch
    │   │   └── workers.go # Connect() - starts asynq Server, polls jobs table, dispatches to handlers
    │   ├── services/     # Business logic
    │   │   └── d2l.go    # D2L API client
    │   ├── util/         # Utility functions
    │   │   ├── job.go    # CreateJob, UpdateJob, GetJob
    │   │   ├── agent.go  # RunAgent wrapper
    │   │   └── s3.go     # S3 helpers
    │   └── routes/       # Route registration: RegisterD2LRoutes, RegisterDevRoutes
    ├── migrations/       # Goose SQL migration files (includes jobs table schema)
    ├── .air.toml         # Air config for main server (hot reload on :8080/4000)
    └── .air.mockapi.toml # Air config for mock API server (hot reload on :4010)
```

### Key patterns

- **Client → Server**: Frontend POSTs to `NEXT_PUBLIC_API_URL` (default `http://localhost:8080` for manual, `http://localhost:4000` in Docker).
    - **D2L API** — `POST /d2l/credentials`, `GET /d2l/courses`, `POST /d2l/sync`
    - **Dev/Agent** — `POST /dev/assignment-files` (upload to S3), `POST /dev/run-agent` (invoke JAJA agent via ADK), `GET /dev/presigned-url` (download URLs)
- **Server → DB**: GORM with PostgreSQL via pgx driver. Global `config.DBClient` initialized via `database.ConnectDB()`. Models include User, Org, D2LCookieSession, D2LLocalStorageSession, and Job (tracks agent job status/results).
- **Server → Redis/asynq**: Asynq job queue (`github.com/hibiken/asynq`) for background task processing. Global `config.RedisClient` initialized via `queue.ConnectRedis()` from `REDIS_URL` env var. Separate from the jobs table (DB-backed): DB polls jobs table, dispatches to handlers via asynq workers.
- **Server → Jobs/Workers**: DB-backed job queue (`internal/models/jobs.go` + `internal/workers/workers.go`). Jobs table tracks status (pending/running/done/failed), type, payload, result. `workers.Server` polls pending jobs, dispatches to handlers in `internal/jobs/handlers/` based on job type (currently JobTypeDocx → HandleDocx).
- **Server → S3**: AWS SDK Go v2 with MinIO (S3-compatible). Global `config.S3BasicsBucket` (BucketBasics struct) initialized via `storage.ConnectObjectStorage()` with static credentials. Helpers in `internal/storage/s3.go` provide CRUD, multipart uploads/downloads, copy, list, exists, presigned URLs. Connection validates via ListBuckets on startup. Presigned URLs signed with `MINIO_PUBLIC_URL` (external via cloudflared) or `MINIO_URL` (internal).
- **Server → Claude AI Agent**: Anthropic SDK Go client (`github.com/anthropics/anthropic-sdk-go`) initialized from `ANTHROPIC_API_KEY` and `ANTHROPIC_API_URL` (defaults to `https://api.anthropic.com`). `agent/runner/runner.go` runs the JAJA agent (Google ADK): takes assignment key, fetches from S3, runs LLM agent, generates `.docx` via `agent/tools/docx.go` (using unioffice), uploads result to S3, updates DB job status. `agent/agents/orchestrator.go` defines the agent with system prompt and docx tool. `agent/models/anthropic.go` adapts Anthropic SDK to ADK model interface (tool conversion, system prompts, max tokens). For testing, set `ANTHROPIC_API_URL` to mock API endpoint (e.g., `http://localhost:4010`).
- **Mock API Server**: Standalone Gin server in `cmd/mockapi/` that mocks Anthropic API responses for testing and development. Runs independently on port 4010 (or `MOCKAPI_PORT` env var). Includes embedded JSON fixtures in `internal/mockapi/handlers/anthropic/testdata/`. Useful for testing agent behavior without consuming API quota or network calls.
- **CORS**: Server reads `FRONTEND_URL` from env to configure allowed origins.
- **shadcn/ui**: Uses Radix Lyra style with Phosphor Icons. Config in `components.json`.
- **Path aliases**: `@/*` maps to project root in TypeScript.

### Environment variables

- Root `.env`: Used by docker-compose services
    - `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` — PostgreSQL credentials
    - `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` — MinIO S3 credentials
- `apps/server/.env`: Used by Gin server and mock API
    - `PORT` — Server port (default 8080 for manual, Docker exposes 4000)
    - `MOCKAPI_PORT` — Mock API port (default 4010, used by docker-compose)
    - `ANTHROPIC_API_URL` — Anthropic API endpoint (default `https://api.anthropic.com`, set to `http://mock-api:4010` or `http://localhost:4010` to use mock server)
    - `FRONTEND_URL` — Client origin for CORS (e.g., http://localhost:3000)
    - `DB_URL` — PostgreSQL DSN (e.g., postgres://user:pass@db:5432/jaja)
    - `REDIS_URL` — Redis endpoint (e.g., redis:6379 in Docker, localhost:6379 for manual runs). **Required** — server fatally exits on startup if unset.
    - `MINIO_URL` — MinIO S3 endpoint (e.g., http://minio:9000)
    - `MINIO_PUBLIC_URL` — Public MinIO S3 endpoint for presigned URLs (e.g., cloudflared tunnel URL). If unset, falls back to `MINIO_URL`.
    - `AWS_REGION` — S3 region (e.g., us-east-1)
    - `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` — S3 credentials (same as root .env)
    - `ANTHROPIC_API_KEY` — Anthropic API key for Claude AI integration (required for agent features)
- `apps/client/.env`: Used by Next.js
    - `NEXT_PUBLIC_API_URL` — Server API endpoint (e.g., http://localhost:8080)

## Key Dependencies

**Go Backend:**
- `github.com/anthropics/anthropic-sdk-go` (v1.30.0) — Official Anthropic SDK for Claude AI integration
- `google.golang.org/adk` — Google Agent Development Kit for building agentic applications
- `google.golang.org/genai` — Google GenAI abstraction layer
- `github.com/hibiken/asynq` (v0.26.0) — Job queue and task scheduler backed by Redis
- `github.com/gin-gonic/gin` — HTTP web framework
- `gorm.io/gorm` + `gorm.io/driver/postgres` — ORM + PostgreSQL driver
- `github.com/aws/aws-sdk-go-v2` — AWS SDK for S3/MinIO operations

## Conventions

- Commit messages use conventional commits format (`feat:`, `dev:`, `fix:`, `refactor:`)
- PRs require a demo video (Loom), test plan, and deployment steps per the PR template
- Dockerfiles are development-only (production builds are TODO)
- The server currently uses a hardcoded test user ID (proper auth is TODO)
- Agent integration is functional: `agent/runner/runner.go` executes the JAJA agent end-to-end (MVP docx generation)

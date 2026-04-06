# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

JAJA (Just Automate Junk Assignments) — a web app for saving D2L (Desire2Learn) cookies and local storage to a database. Monorepo with a Next.js frontend and Go backend, backed by PostgreSQL, MinIO (S3-compatible object storage), and Claude AI (Anthropic) for assignment completion.

## Development Commands

### Full stack (Docker)

```bash
# Start cloudflared tunnel on port 9000 (in a separate terminal)
cloudflared tunnel --url http://localhost:9000

# Then start the full stack
docker compose up          # Starts db, minio, server, and client with hot reload
```

- Frontend: http://localhost:3000
- Server: http://localhost:4000 (via Docker) or http://localhost:8080 (manual)
- PostgreSQL: localhost:5432
- MinIO API: http://localhost:9000
- MinIO Console: http://localhost:9001

Note: The cloudflared tunnel URL (from the first command) should be set as `MINIO_PUBLIC_URL` in `apps/server/.env` for presigned URLs that Claude AI can access.

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

This starts the server on port specified by `PORT` env var (default 8080). The Dockerfile exposes port 4000 and uses Air for live reloading.

### Agent web UI (experimental)

To run the experimental Claude AI agent with an interactive web interface:

```bash
cd apps/server/agent
go run agent.go web webui api
```

This launches a chat interface where you can interact with the agent directly. The web UI provides debugging information and dev logs for development and testing purposes. This is experimental functionality and currently uses boilerplate setup from Google ADK.

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
└── server/               # Go 1.25 + Gin + GORM + Anthropic SDK
    ├── cmd/main.go       # Entry point: loads env, connects DB + S3, sets up router
    ├── agent/            # Claude AI integration via Anthropic SDK
    │   ├── agent.go      # Agent setup using Google ADK (experimental)
    │   └── models/       # AnthropicModel wrapper for ADK compatibility
    │       └── anthropic.go # Anthropic SDK client with ADK model interface
    ├── internal/
    │   ├── config/       # ConnectDB(), ConnectObjectStorage() (global vars: DBClient, S3BasicsBucket)
    │   ├── handlers/
    │   │   ├── d2l/      # D2L handlers: SaveCredentials, GetCoursesAndAssignments, SyncCoursesAndAssignments
    │   │   └── dev/      # Dev handlers: SaveAssignmentFiles (CompleteAssignment WIP)
    │   ├── models/       # GORM models: User, Org, D2LCookieSession, D2LLocalStorageSession
    │   ├── storage/      # BucketBasics S3 operations (CRUD, multipart, presigned URLs)
    │   ├── services/     # Business logic (d2l.go for D2L API client)
    │   └── routes/       # Route registration: RegisterD2LRoutes, RegisterDevRoutes
    └── migrations/       # Goose SQL migration files
```

### Key patterns

- **Client → Server**: Frontend POSTs to `NEXT_PUBLIC_API_URL` (default `http://localhost:8080` for manual, `http://localhost:4000` in Docker).
    - `POST /d2l/credentials` — Save D2L cookies & localStorage (form data)
    - `GET /d2l/courses` — Load user's courses and assignments from D2L
    - `POST /d2l/sync` — Sync courses and assignments from D2L to database
    - `POST /dev/assignment-files` — Upload assignment files to S3 (multipart form data)
    - `POST /dev/complete-assignment` — Submit assignment to Claude AI for completion (currently disabled/WIP)
- **Server → DB**: GORM with PostgreSQL via pgx driver. Global `config.DBClient` variable initialized via `config.ConnectDB()`. Used across handlers via `config.DBClient.Create()`, `.Query()`, etc.
- **Server → S3**: AWS SDK Go v2 with MinIO (S3-compatible). Global `config.S3BasicsBucket` (BucketBasics struct) initialized via `config.ConnectObjectStorage()` with static credentials. Provides bucket/object operations (CRUD, multipart uploads/downloads, copy, list, exists check, presigned URLs). Connection validates via ListBuckets on startup. Presigned URLs are signed with `MINIO_PUBLIC_URL` (for external access via cloudflared) or `MINIO_URL` (internal).
- **Server → Claude AI**: Anthropic SDK Go client (`github.com/anthropics/anthropic-sdk-go`) initialized from `ANTHROPIC_API_KEY` env var. `agent/models/anthropic.go` provides an AnthropicModel wrapper that implements the Google ADK model interface for compatibility with agent frameworks. `CompleteAssignment` handler (currently WIP) will generate presigned S3 URLs and send PDFs to Claude for processing.
- **CORS**: Server reads `FRONTEND_URL` from env to configure allowed origins.
- **shadcn/ui**: Uses Radix Lyra style with Phosphor Icons. Config in `components.json`.
- **Path aliases**: `@/*` maps to project root in TypeScript.

### Environment variables

- Root `.env`: Used by docker-compose services
    - `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` — PostgreSQL credentials
    - `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` — MinIO S3 credentials
- `apps/server/.env`: Used by Gin server
    - `PORT` — Server port (default 8080 for manual, Docker exposes 4000)
    - `FRONTEND_URL` — Client origin for CORS (e.g., http://localhost:3000)
    - `DB_URL` — PostgreSQL DSN (e.g., postgres://user:pass@db:5432/jaja)
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
- `github.com/gin-gonic/gin` — HTTP web framework
- `gorm.io/gorm` + `gorm.io/driver/postgres` — ORM + PostgreSQL driver
- `github.com/aws/aws-sdk-go-v2` — AWS SDK for S3/MinIO operations

## Conventions

- Commit messages use conventional commits format (`feat:`, `dev:`, `fix:`)
- PRs require a demo video (Loom), test plan, and deployment steps per the PR template
- Dockerfiles are development-only (production builds are TODO)
- The server currently uses a hardcoded test user ID (proper auth is TODO)
- Agent integration is WIP: `agent.go` contains experimental boilerplate using Google ADK

# jaja

JAJA: Just Automate Junk Assignments — a web app for saving D2L (Desire2Learn) cookies and local storage to a database, with Claude AI agent integration for academic assignment completion. Monorepo with a Next.js frontend and Go backend, backed by PostgreSQL, MinIO (S3-compatible object storage), Redis (job queue), and Anthropic Claude AI (via Google Agent Development Kit).

## Prerequisites

- [Docker](https://www.docker.com/) (recommended for full stack)
- [Bun](https://bun.sh) (frontend)
- [Go](https://go.dev) 1.25+ (backend)

## Quick Start (Docker)

```bash
# Terminal 1: Start Docker services
cp .env.example .env
docker compose up

# Terminal 2: Start the frontend
cd apps/client && bun dev
```

| Service       | URL                   |
| ------------- | --------------------- |
| Frontend      | http://localhost:3000 |
| Server API    | http://localhost:4000 |
| PostgreSQL    | localhost:5432        |
| Redis         | localhost:6379        |
| MinIO API     | http://localhost:9000 |
| MinIO Console | http://localhost:9001 |

**Note:** The frontend must be started separately with `bun dev` in a second terminal. Docker Compose only starts the backend services (db, redis, minio, server).

## Setup

### Frontend (`apps/client/`)

```bash
cd apps/client
bun install
bunx skills add vercel-labs/next-browser  # AI agent debugging tools (installs Chromium via Playwright)
bun dev
```

Runs at http://localhost:3000

### Backend (`apps/server/`)

The server runs via Air (hot reload) inside Docker. To run manually:

```bash
cd apps/server
cp .env.example .env        # configure DB_URL, MINIO_URL, etc.
go mod download
go run cmd/main.go
```

Runs at http://localhost:8080

### Environment Variables

Copy the example files and fill in your values:

- **Root** `.env` — Docker Compose services (`POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`)
- **Server** `apps/server/.env` — Go server (`PORT`, `FRONTEND_URL`, `DB_URL`, `REDIS_URL`, `MINIO_URL`, `MINIO_PUBLIC_URL`, `AWS_REGION`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, `ANTHROPIC_API_KEY`)
- **Client** `apps/client/.env` — Next.js (`NEXT_PUBLIC_API_URL`)

### ADK Agent Dev Tools (`apps/server/`)

`cmd/agent/main.go` is a standalone dev entrypoint for interacting with the JAJA agents outside of Docker. It supports two modes.

```bash
# apps/server/.env.local
MINIO_URL=http://localhost:9000
```

#### Console mode (terminal)

Runs an interactive chat session with the agent directly in your terminal — no browser needed.

```bash
cd apps/server
go run cmd/agent/main.go console
```
Can also be run by executing agent.exe

Type your prompt and press Enter. The agent responds inline. `Ctrl+C` to exit.

#### Web UI mode (browser)

Starts a local server with a browser UI for chatting with and debugging the agents.

```bash
cd apps/server
go run cmd/agent/main.go web webui api
```

Open http://localhost:8080. Use the agent dropdown to switch between `jaja_orchestrator`, `analysis_agent`, and `docx_agent`.

> `MINIO_PUBLIC_URL` in `apps/server/.env` must be a publicly accessible URL (e.g. a Cloudflare tunnel to `localhost:9000`) so presigned S3 URLs are reachable by Claude AI.

## API Endpoints

### D2L Integration

- `POST /d2l/credentials` — Save D2L cookies and localStorage data
- `GET /d2l/courses` — Load user's courses and assignments from D2L
- `POST /d2l/sync` — Sync courses and assignments from D2L to database

### Agent & Assignment Completion

- `POST /dev/assignment-files` — Upload assignment files (PDF, instructions, rubric) to S3 storage
- `POST /dev/run-agent` — Invoke the JAJA agent to analyze and complete an assignment (generates `.docx` Word document)
- `POST /dev/run-claude` — Direct Claude API call for assignment completion (non-agent path)
- `GET /dev/presigned-url` — Generate presigned S3 URLs for downloading assignment files and results

## Tech Stack

- **Frontend**: Next.js 16 (App Router), React 19, Tailwind CSS 4, shadcn/ui, TypeScript
- **Backend**: Go 1.25, Gin, GORM, asynq (Redis-backed job queue), Google ADK (agent framework)
- **Database**: PostgreSQL (schema managed by GORM + Goose migrations)
- **Job Queue**: Redis + asynq (asynq workers) + DB jobs table (polling-based job dispatch)
- **Object Storage**: MinIO (S3-compatible, AWS SDK Go v2)
- **AI**: Anthropic Claude API (via `anthropic-sdk-go`), Google Agent Development Kit (ADK) for orchestration
- **Document Generation**: unioffice (Go library for `.docx` Word document creation)

# jaja

JAJA: Just Automate Junk Assignments — a web app for saving D2L (Desire2Learn) cookies and local storage to a database, with Claude AI integration for assignment completion. Monorepo with a Next.js frontend and Go backend, backed by PostgreSQL and MinIO (S3-compatible object storage).

## Prerequisites

- [Docker](https://www.docker.com/) (recommended for full stack)
- [Bun](https://bun.sh) (frontend)
- [Go](https://go.dev) 1.25+ (backend)

## Quick Start (Docker)

```bash
cp .env.example .env
docker compose up
```

| Service       | URL                   |
| ------------- | --------------------- |
| Frontend      | http://localhost:3000 |
| Server API    | http://localhost:4000 |
| PostgreSQL    | localhost:5432        |
| MinIO API     | http://localhost:9000 |
| MinIO Console | http://localhost:9001 |

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
- **Server** `apps/server/.env` — Go server (`PORT`, `FRONTEND_URL`, `DB_URL`, `MINIO_URL`, `MINIO_PUBLIC_URL`, `AWS_REGION`, `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD`, `ANTHROPIC_API_KEY`)
- **Client** `apps/client/.env` — Next.js (`NEXT_PUBLIC_API_URL`)

## API Endpoints

### D2L Integration

- `POST /d2l/credentials` — Save D2L cookies and localStorage data
- `GET /d2l/courses` — Load user's courses and assignments from D2L
- `POST /d2l/sync` — Sync courses and assignments from D2L to database

### Dev (Development/Testing)

- `POST /dev/assignment-files` — Upload assignment files (instructions/rubric) to S3 storage
- `POST /dev/complete-assignment` — Submit assignment to Claude AI for completion (WIP)

## Tech Stack

- **Frontend**: Next.js 16 (App Router), React 19, Tailwind CSS 4, shadcn/ui, TypeScript
- **Backend**: Go 1.25, Gin, GORM
- **Database**: PostgreSQL
- **Object Storage**: MinIO (S3-compatible)
- **AI**: Anthropic Claude API, Google Agent Development Kit (experimental)

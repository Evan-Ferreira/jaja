# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

JAJA (Just Automate Junk Assignments) — a web app for saving D2L (Desire2Learn) cookies and local storage to a database. Monorepo with a Next.js frontend and Go backend, backed by PostgreSQL and MinIO (S3-compatible object storage).

## Development Commands

### Full stack (Docker)
```bash
docker compose up          # Starts db, minio, server, and client with hot reload
```
- Frontend: http://localhost:3000
- Server: http://localhost:8080
- PostgreSQL: localhost:5432
- MinIO API: http://localhost:9000
- MinIO Console: http://localhost:9001

### Frontend only (`apps/client/`)
```bash
bun install
bun dev                    # Next.js dev server on :3000
bun build                  # Production build
bunx eslint .              # Lint
```

### Backend only (`apps/server/`)
The server runs via Air (hot reload) inside Docker. To run manually:
```bash
go mod download
go run cmd/main.go
```

### Database migrations (Goose)
Migrations live in `apps/server/migrations/` using Goose SQL format. GORM auto-migrates models on startup, but schema changes should also have corresponding Goose migration files. Environment variables for Goose are in the root `.env`.

## Architecture

```
apps/
├── client/               # Next.js 16 (App Router) + React 19 + Tailwind 4 + shadcn/ui
│   ├── app/              # Pages and app-level components
│   │   └── components/   # Page-specific components (e.g., cookie-form.tsx)
│   ├── components/ui/    # shadcn UI primitives (button, checkbox, field, etc.)
│   ├── lib/utils.ts      # cn() helper (clsx + tailwind-merge)
│   └── utils/string.ts   # parseStringToJSON() for tab-separated cookie/storage data
└── server/               # Go 1.25 + Gin + GORM
    ├── cmd/main.go       # Entry point: loads env, connects DB + S3, sets up router
    ├── internal/
    │   ├── database/     # ConnectDB() (global DBClient var)
    │   ├── storage/      # ConnectObjectStorage() (global S3Client), BucketBasics S3 operations
    │   ├── models/       # GORM models: User, D2LCookieSession, D2LLocalStorageSession
    │   └── routes/d2l/   # Route handlers: POST /api/d2l/auth
    └── migrations/       # Goose SQL migration files
```

### Key patterns
- **Client → Server**: Frontend POSTs to `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`). Single API endpoint: `POST /api/d2l/auth` accepts `{ cookies, local_storage }`.
- **Server → DB**: GORM with PostgreSQL via pgx driver. Global `database.DBClient` variable initialized via `database.ConnectDB()`. Used across handlers via `database.DBClient.Create()`, `.Query()`, etc.
- **Server → S3**: AWS SDK Go v2 with MinIO (S3-compatible). Global `storage.S3Client` initialized via `storage.ConnectObjectStorage()` with static credentials. `BucketBasics` struct provides bucket/object operations (CRUD, multipart uploads/downloads, copy, list). Connection validates via ListBuckets on startup.
- **CORS**: Server reads `FRONTEND_URL` from env to configure allowed origins.
- **shadcn/ui**: Uses Radix Lyra style with Phosphor Icons. Config in `components.json`.
- **Path aliases**: `@/*` maps to project root in TypeScript.

### Environment variables
- Root `.env`: Used by docker-compose services
  - `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB` — PostgreSQL credentials
  - `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` — MinIO S3 credentials
- `apps/server/.env`: Used by Gin server
  - `PORT` — Server port (default 8080)
  - `FRONTEND_URL` — Client origin for CORS (e.g., http://localhost:3000)
  - `DB_URL` — PostgreSQL DSN (e.g., postgres://user:pass@db:5432/jaja)
  - `MINIO_URL` — MinIO S3 endpoint (e.g., http://minio:9000)
  - `AWS_REGION` — S3 region (e.g., us-east-1)
  - `MINIO_ROOT_USER`, `MINIO_ROOT_PASSWORD` — S3 credentials (same as root .env)
- `apps/client/.env`: Used by Next.js
  - `NEXT_PUBLIC_API_URL` — Server API endpoint (e.g., http://localhost:8080)

## Conventions

- Commit messages use conventional commits format (`feat:`, `dev:`, `fix:`)
- PRs require a demo video (Loom), test plan, and deployment steps per the PR template
- Dockerfiles are development-only (production builds are TODO)
- The server currently uses a hardcoded test user ID (proper auth is TODO)

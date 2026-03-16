# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

JAJA (Just Automate Junk Assignments) — a web app for saving D2L (Desire2Learn) cookies and local storage to a database. Monorepo with a Next.js frontend and Go backend, backed by PostgreSQL.

## Development Commands

### Full stack (Docker)
```bash
docker compose up          # Starts db, server, and client with hot reload
```
- Frontend: http://localhost:3000
- Server: http://localhost:8080
- PostgreSQL: localhost:5432

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
    ├── cmd/main.go       # Entry point: loads config, connects DB, sets up router
    ├── internal/
    │   ├── config/       # LoadConfig() (.env), ConnectDB() (global DB var)
    │   ├── models/       # GORM models: User, D2LCookieSession, D2LLocalStorageSession
    │   └── routes/d2l/   # Route handlers: POST /api/d2l/auth
    └── migrations/       # Goose SQL migration files
```

### Key patterns
- **Client → Server**: Frontend POSTs to `NEXT_PUBLIC_API_URL` (default `http://localhost:8080`). Single API endpoint: `POST /api/d2l/auth` accepts `{ cookies, local_storage }`.
- **Server → DB**: GORM with PostgreSQL via pgx driver. Global `config.DB` variable used across handlers.
- **CORS**: Server reads `FRONTEND_URL` from env to configure allowed origins.
- **shadcn/ui**: Uses Radix Lyra style with Phosphor Icons. Config in `components.json`.
- **Path aliases**: `@/*` maps to project root in TypeScript.

### Environment variables
- Root `.env`: PostgreSQL credentials, Goose config
- `apps/server/.env`: `PORT`, `FRONTEND_URL`, `DB_URL`, Goose config
- `apps/client/.env`: `NEXT_PUBLIC_API_URL`

## Conventions

- Commit messages use conventional commits format (`feat:`, `dev:`, `fix:`)
- PRs require a demo video (Loom), test plan, and deployment steps per the PR template
- Dockerfiles are development-only (production builds are TODO)
- The server currently uses a hardcoded test user ID (proper auth is TODO)

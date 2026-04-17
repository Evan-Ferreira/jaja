# JAJA Codemaps Index

**Last Updated:** 2026-04-16

This directory contains architecture maps and design documents for the JAJA (Just Automate Junk Assignments) project. Each codemap describes a subsystem, its components, data flow, and external dependencies.

## Maps

### [Backend Agent & Workers](./AGENT_WORKERS.md)
The Claude AI agent integration and database-backed job queue system. Covers the JAJA agent (Google ADK), agent runner, tools (docx generation), and the workers that poll and dispatch jobs from the DB.

**Key Files:**
- `apps/server/agent/` — Agent orchestration, runner, tools
- `apps/server/internal/workers/` — Job polling and dispatch
- `apps/server/internal/jobs/` — Job types and handlers
- `apps/server/internal/models/jobs.go` — Job schema

### [Backend Infrastructure & Services](./BACKEND.md)
Server configuration, database, storage, and D2L integration. Covers the modular config packages (database, queue, storage), PostgreSQL schema, MinIO S3 integration, and D2L API client.

**Key Files:**
- `apps/server/internal/database/` — PostgreSQL connection
- `apps/server/internal/queue/` — Redis/asynq setup
- `apps/server/internal/storage/` — MinIO S3 operations
- `apps/server/internal/services/` — D2L API client, Claude integration
- `apps/server/migrations/` — Database schema (Goose)

### [Frontend Architecture](./FRONTEND.md)
Next.js app structure, pages, components, and utilities. Covers the main flows (cookie form submission, course sync, dev file upload).

**Key Files:**
- `apps/client/app/` — Pages and layouts
- `apps/client/components/` — React components
- `apps/client/lib/utils.ts` — Helper utilities

---

## Quick Navigation

**Need to:**
- Add a new agent tool? → [AGENT_WORKERS.md](./AGENT_WORKERS.md#tools)
- Add a new API route? → [BACKEND.md](./BACKEND.md#routes)
- Add a new job type? → [AGENT_WORKERS.md](./AGENT_WORKERS.md#job-types)
- Modify database schema? → [BACKEND.md](./BACKEND.md#database-schema)
- Add a new page? → [FRONTEND.md](./FRONTEND.md)

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                     Next.js Frontend                      │
│                  (React 19, Tailwind CSS)                │
└────────────────────────────────────────────────────────┘
                              │
                      ┌───────┴────────┐
                      │                │
                      ▼                ▼
            ┌──────────────────┐  ┌──────────┐
            │   Gin Server     │  │  MinIO   │
            │   (Go 1.25)      │  │   S3     │
            └──────────────────┘  └──────────┘
                      │
       ┌──────────────┼──────────────┐
       │              │              │
       ▼              ▼              ▼
    ┌─────┐     ┌─────────┐    ┌────────┐
    │ DB  │     │  Redis  │    │ Claude │
    │ PG  │     │  asynq  │    │  API   │
    └─────┘     └─────────┘    └────────┘
       │              │
       └──────────────┴─────────────────┬──────────────┐
                                        │              │
                                        ▼              ▼
                            ┌─────────────────┐  ┌──────────────┐
                            │  DB Job Queue   │  │  ADK Agent   │
                            │  (polling)      │  │  (Docx Tool) │
                            └─────────────────┘  └──────────────┘
```

## Environment & Setup

See `/CLAUDE.md` for:
- Development commands (Docker, manual)
- Environment variables
- Dependency versions
- Database migrations

## Related Documentation

- **/.claude/CLAUDE.md** — Developer guide, commands, patterns
- **/README.md** — Project overview, quick start
- **/docs/auth/** — D2L SSO and authentication docs

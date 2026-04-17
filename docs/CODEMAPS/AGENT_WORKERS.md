# Agent & Workers Codemap

**Last Updated:** 2026-04-16  
**Status:** MVP (docx generation working, extensible for other job types)

## Overview

The JAJA agent is a Claude AI-powered assignment completion agent built on Google's Agent Development Kit (ADK). It analyzes assignment files, generates a response, and outputs a submission-ready Word document (`.docx`).

The workers system is a database-backed job queue that polls a `jobs` table for pending work and dispatches to handlers based on job type. Separate from Redis/asynq for reliability and horizontal scaling.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                    HTTP Handler Layer                         │
│  POST /dev/run-agent   POST /dev/run-claude                  │
└───────────────────┬────────────────────────┬──────────────────┘
                    │                        │
        ┌───────────▼─────────┐  ┌──────────▼──────────┐
        │  Job Creation       │  │  Direct Claude Call │
        │  (internal/util)    │  │  (internal/services)│
        └───────────┬─────────┘  └─────────────────────┘
                    │
        ┌───────────▼──────────────────┐
        │   DB: jobs table INSERT      │
        │   (status=pending)           │
        └───────────┬──────────────────┘
                    │
        ┌───────────▼────────────────────────┐
        │  Workers: polling loop             │
        │  (internal/workers/workers.go)     │
        │  → SELECT * FROM jobs WHERE ...    │
        │  → Dispatch by job.Type            │
        └───────────┬────────────────────────┘
                    │
        ┌───────────▼─────────────────────┐
        │  Job Handlers                   │
        │  (internal/jobs/handlers/)      │
        │  - HandleDocx                   │
        │  - HandleUnknown                │
        └───────────┬─────────────────────┘
                    │
        ┌───────────▼──────────────────────────┐
        │  Agent Runner                        │
        │  (agent/runner/runner.go)            │
        │  → Init agent                        │
        │  → Run LLM loop                      │
        │  → Call tools (docx)                 │
        │  → Upload result to S3               │
        │  → Update job status                 │
        └──────────────────────────────────────┘
```

## Key Modules

| Module | Purpose | Key Files | Responsibility |
|--------|---------|-----------|-----------------|
| **Agent Orchestration** | Define JAJA agent with tools | `agent/agents/orchestrator.go` | `CreateJAJAAgent()` creates named agent with docx tool, system prompt |
| **Agent Runner** | Execute agent end-to-end | `agent/runner/runner.go` | Handles LLM loop, tool dispatch, result upload to S3 |
| **Agent Tools** | Tool implementations for agent | `agent/tools/docx.go` | `create_docx` tool using unioffice for `.docx` generation |
| **Agent Models** | Anthropic SDK adapter for ADK | `agent/models/anthropic.go` | Wraps Anthropic SDK with ADK model interface (tools, prompts, max tokens) |
| **Agent Config** | Global agent initialization | `agent/config.go` | `ConnectAgent()` initializes global `AgentRunner` |
| **Workers** | Job polling and dispatch | `internal/workers/workers.go` | Polls DB for pending jobs, dispatches to handlers by type |
| **Job Handlers** | Task processors | `internal/jobs/handlers/*.go` | Handler funcs: `HandleDocx`, `HandleUnknown` |
| **Job Types** | Job type constants | `internal/jobs/types.go` | Defines `JobTypeDocx = "docx"` etc. |
| **Job Model** | Job schema in DB | `internal/models/jobs.go` | GORM model: id, type, payload, state, result, timestamps |
| **Job Utils** | Job CRUD operations | `internal/util/job.go` | `CreateJob()`, `UpdateJob()`, `GetJob()` |

## Data Flow

### 1. Trigger: POST /dev/run-agent

```go
{
  "session_id": "user-session-123",
  "prompt": "Complete this essay on climate change...",
  "user_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

**Handler:** `internal/handlers/dev/run-agent.go` → `RunAgent()`
- Validates request (session_id, prompt required)
- Calls `internal/util.CreateJob()` to insert pending job in DB
- **Returns:** Job ID to client immediately (async)

### 2. Worker Polling: internal/workers/workers.go

```sql
SELECT id, type, payload, state, created_at
FROM jobs
WHERE state = 'pending'
ORDER BY created_at ASC
LIMIT 10
```

**Process:**
1. asynq Server calls registered handler by job.Type
2. Handler retrieves full job from DB
3. Dispatches to processor (e.g., HandleDocx)

### 3. Job Processing: internal/jobs/handlers/docx.go

```go
type DocxJobPayload struct {
  UserID        string `json:"user_id"`
  AssignmentKey string `json:"assignment_key"`
  Prompt        string `json:"prompt"`
}
```

**Flow:**
1. Deserialize job.Payload
2. Fetch assignment file from S3 (via AssignmentKey)
3. Call `agent.AgentRunner.Run(ctx, RunInput{...})`
4. Receive `.docx` bytes from agent
5. Upload result to S3 with key `{AssignmentKey}_result.docx`
6. Update job in DB: state='done', result=docx_bytes
7. On error: state='failed', last_err set, retried incremented

### 4. Agent Execution: agent/runner/runner.go

**RunInput:**
```go
type RunInput struct {
  SessionID string  // ADK session ID for conversation history
  Prompt    string  // Assignment description + instructions
  UserID    string  // Owner for audit/multi-tenancy
}
```

**Agent Run Loop:**
1. Send user prompt to JAJA agent
2. LLM responds with function call to `create_docx` tool
3. Tool execution: unioffice generates `.docx`
4. Return docx bytes to runner
5. Runner uploads to S3, updates job, returns result path

## Tools

### create_docx

**Purpose:** Generate submission-ready Word document from assignment response

**Implementation:** `agent/tools/docx.go`

**Signature:**
```go
func DocxTool(ctx context.Context, input struct {
  Title    string `json:"title"`
  Content  string `json:"content"`
  Sections []struct {
    Heading string
    Body    string
  } `json:"sections"`
}) ([]byte, error)
```

**Returns:** `.docx` file bytes (written to S3 by runner)

**Libraries:** `github.com/unidoc/unioffice` (unidoc package)

**Extensibility:**
- Add more formatting (styles, images, tables) to `unioffice` calls
- Add new tool: `research_tool`, `cite_tool`, etc. in `agent/tools/` + register in `agents/orchestrator.go`

## Job Types

Currently defined in `internal/jobs/types.go`:

```go
const (
  JobTypeDocx = "docx"
)
```

**Adding a new job type:**

1. Add constant to `internal/jobs/types.go`
2. Define payload struct in `internal/jobs/handlers/new_handler.go`
3. Implement handler func: `HandleNewType(ctx context.Context, t *asynq.Task) error`
4. Register in `internal/workers/workers.go`: `mux.HandleFunc(JobTypeNewType, HandleNewType)`
5. Create corresponding tool if it's an agent job

## Database Schema

**jobs table** (created by migration `20260415025611_add_jobs_table.sql`):

```sql
CREATE TABLE jobs (
  id TEXT PRIMARY KEY,
  queue TEXT NOT NULL DEFAULT 'default',
  type TEXT NOT NULL,
  payload JSONB,
  state TEXT NOT NULL,
  max_retry INT DEFAULT 25,
  retried INT DEFAULT 0,
  last_err TEXT,
  last_failed_at TIMESTAMPTZ,
  result BYTEA,
  completed_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now(),
  updated_at TIMESTAMPTZ DEFAULT now()
);
```

**State Transitions:**

```
pending → running → done
       ↘          ↗
         failed (retry if retried < max_retry)
```

## Related Areas

- **Frontend:** Calls `POST /dev/run-agent` from dev pages
- **Backend Routes:** `internal/routes/dev.go` registers endpoints
- **Storage:** S3 integration in `internal/storage/` for file I/O
- **Database:** Job persistence in PostgreSQL via GORM
- **Queue:** Redis/asynq integration in `internal/queue/` for worker execution

## Dependencies

**Go Packages:**
- `google.golang.org/adk` (v0.x) — Agent framework
- `google.golang.org/genai` — GenAI abstractions
- `github.com/anthropics/anthropic-sdk-go` (v1.30.0) — Claude API
- `github.com/hibiken/asynq` (v0.26.0) — Async task processing
- `github.com/unidoc/unioffice` — Word document generation
- `gorm.io/gorm` — ORM for job persistence

**Environment Variables:**
- `ANTHROPIC_API_KEY` — Required for agent initialization
- `REDIS_URL` — Required for asynq workers
- `DB_URL` — Required for job polling
- `MINIO_URL`, `MINIO_PUBLIC_URL` — S3 for file storage

## Testing

### Manual Testing

```bash
# 1. Start server with agent + workers
docker compose up

# 2. Create a job (trigger via curl)
curl -X POST http://localhost:8080/dev/run-agent \
  -H "Content-Type: application/json" \
  -d '{
    "session_id": "test-session",
    "prompt": "Write a 500-word essay on climate change",
    "user_id": "550e8400-e29b-41d4-a716-446655440000"
  }'

# Returns: {"response": "job-id-uuid"}

# 3. Poll job status
curl http://localhost:8080/dev/job-status/{job-id}

# 4. Download result
curl http://localhost:8080/dev/presigned-url?path=result.docx
```

### Unit Tests

Located in `apps/server/internal/jobs/handlers/` (to be implemented).

## Future Enhancements

1. **Streaming:** Real-time agent response streaming via WebSocket
2. **Multi-Tool Workflows:** Orchestrate research → write → cite tools in sequence
3. **Human-in-the-Loop:** Pause agent for review/approval before final submission
4. **Result Caching:** Store common assignment types for faster completion
5. **Cost Optimization:** Token counting, batch job processing, model selection per complexity
6. **Monitoring:** Job metrics, error tracking, performance dashboards

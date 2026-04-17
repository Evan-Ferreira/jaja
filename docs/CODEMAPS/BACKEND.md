# Backend Infrastructure & Services Codemap

**Last Updated:** 2026-04-16

## Overview

The Go backend handles API routing, database operations, external integrations (D2L, Anthropic Claude), and object storage (MinIO). It's organized into modular config packages (database, queue, storage) with centralized initialization in `cmd/main.go`.

## Architecture

```
┌──────────────────────────────────────────────────────────┐
│                    cmd/main.go                            │
│  ┌────────────────────────────────────────────────────┐  │
│  │  Load .env                                          │  │
│  │  → database.ConnectDB()                             │  │
│  │  → queue.ConnectRedis()                             │  │
│  │  → storage.ConnectObjectStorage()                   │  │
│  │  → agent.ConnectAgent()                             │  │
│  │  → workers.Connect()                                │  │
│  │  → Setup Gin router + CORS                          │  │
│  │  → Register routes                                  │  │
│  └────────────────────────────────────────────────────┘  │
└────────────┬─────────────────────────────────────────────┘
             │
   ┌─────────┴──────────┬──────────────┬──────────────┐
   │                    │              │              │
   ▼                    ▼              ▼              ▼
┌────────────┐  ┌────────────┐  ┌─────────┐  ┌──────────┐
│ Database   │  │ Queue      │  │ Storage │  │ Agent    │
│ (PG)       │  │ (Redis)    │  │ (MinIO) │  │ (Claude) │
└────────────┘  └────────────┘  └─────────┘  └──────────┘
   │                    │              │              │
   ▼                    ▼              ▼              ▼
┌─────────────────────────────────────────────────────┐
│                  Gin HTTP Router                     │
│  GET /d2l/courses, POST /d2l/sync, etc.            │
│  POST /dev/assignment-files, /dev/run-agent, etc.  │
└─────────────────────────────────────────────────────┘
```

## Key Modules

| Module | Purpose | Key Files | Responsibility |
|--------|---------|-----------|-----------------|
| **Database Config** | PostgreSQL connection | `internal/database/config.go` | `ConnectDB()` initializes GORM client via `DB_URL` |
| **Queue Config** | Redis + asynq setup | `internal/queue/config.go` | `ConnectRedis()` initializes RedisClient for worker dispatch |
| **Storage Config** | MinIO S3 initialization | `internal/storage/config.go` | `ConnectObjectStorage()` initializes AWS SDK v2 client |
| **Storage S3 Ops** | S3 CRUD operations | `internal/storage/s3.go` | BucketBasics helpers: PutObject, GetObject, ListObjects, PresignedURL |
| **D2L Service** | D2L API client | `internal/services/d2l.go` | Authenticates with D2L, fetches courses/assignments |
| **Claude Service** | Claude integration | `internal/services/claude.go` | Direct Claude API calls, assignment completion logic |
| **D2L Handlers** | D2L endpoints | `internal/handlers/d2l/*.go` | SaveCredentials, GetCoursesAndAssignments, SyncCoursesAndAssignments |
| **Dev Handlers** | Testing endpoints | `internal/handlers/dev/*.go` | SaveAssignmentFiles, RunAgent, RunClaude, GeneratePresignedURL |
| **Routes** | Route registration | `internal/routes/*.go` | RegisterD2LRoutes, RegisterDevRoutes |
| **Models** | GORM schemas | `internal/models/*.go` | User, Org, D2LCookieSession, D2LLocalStorageSession, Job |

## Config Package Organization

### Before (Monolithic)
```
internal/config/
  ├── config.go
  ├── database.go
  ├── redis.go
  ├── storage.go
  └── workers.go
```

### After (Modular)
```
internal/
  ├── config.go              # Global vars (DBClient, RedisClient, etc.)
  ├── database/
  │   └── config.go          # ConnectDB()
  ├── queue/
  │   └── config.go          # ConnectRedis()
  ├── storage/
  │   ├── config.go          # ConnectObjectStorage()
  │   └── s3.go              # BucketBasics operations
  └── workers/
      └── workers.go         # Connect(), Server startup
```

## Global Variables

Initialized in `cmd/main.go` and used across handlers:

```go
// internal/config.go
var (
  DBClient       *gorm.DB              // PostgreSQL connection
  RedisClient    *asynq.Client         // Redis client for task enqueueing
  S3BasicsBucket storage.BucketBasics  // MinIO S3 operations
  // Agent and Workers initialized in agent.ConnectAgent() + workers.Connect()
)

// agent/config.go
var (
  AgentRunner *runner.AgentRunner  // Global agent runner instance
)

// internal/workers/workers.go
var (
  Server *asynq.Server  // Asynq worker server (polls jobs table)
)
```

## Database Schema

**Models defined in `internal/models/`:**

1. **User** — Stores user identity for multi-tenancy
   ```go
   type User struct {
     ID        uuid.UUID
     Email     string
     CreatedAt time.Time
     UpdatedAt time.Time
   }
   ```

2. **Org** — Organization/institution
   ```go
   type Org struct {
     ID   uuid.UUID
     Name string
   }
   ```

3. **D2LCookieSession** — D2L cookie storage
   ```go
   type D2LCookieSession struct {
     ID        string // UUID
     UserID    uuid.UUID
     Cookies   []byte // JSON
     ExpiresAt time.Time
   }
   ```

4. **D2LLocalStorageSession** — D2L localStorage storage
   ```go
   type D2LLocalStorageSession struct {
     ID        string // UUID
     UserID    uuid.UUID
     Data      []byte // JSON
     ExpiresAt time.Time
   }
   ```

5. **Job** — Assignment completion job
   ```go
   type Job struct {
     ID          string    // UUID
     Type        string    // "docx"
     Payload     []byte    // JSON
     State       string    // "pending", "running", "done", "failed"
     Result      []byte    // Docx bytes or error
     CreatedAt   time.Time
     UpdatedAt   time.Time
   }
   ```

**Migrations:**
- Located in `apps/server/migrations/` (Goose format)
- `20260415025611_add_jobs_table.sql` — Latest migration for jobs table
- GORM auto-migrates on startup, but schema changes should have corresponding Goose files

## Routes

### D2L Integration Routes

**File:** `internal/routes/d2l.go` + `internal/handlers/d2l/`

```
POST /d2l/credentials          → SaveCredentials
  ├─ Form: cookies (string), localStorage (string)
  └─ Stores encrypted sessions in DB

GET /d2l/courses               → GetCoursesAndAssignments
  ├─ Param: org_id (optional)
  └─ Fetches from D2L API

POST /d2l/sync                 → SyncCoursesAndAssignments
  ├─ Body: course IDs, assignment IDs
  └─ Syncs to local DB
```

### Dev/Assignment Routes

**File:** `internal/routes/dev.go` + `internal/handlers/dev/`

```
POST /dev/assignment-files     → SaveAssignmentFiles
  ├─ Multipart form: files (PDF, images)
  └─ Uploads to S3 (MinIO)

POST /dev/run-agent            → RunAgent
  ├─ Body: {session_id, prompt, user_id}
  └─ Creates job, returns job ID

POST /dev/run-claude           → RunClaude
  ├─ Body: {prompt, files}
  └─ Direct Claude call (non-agent)

GET /dev/presigned-url         → GeneratePresignedURL
  ├─ Query: path (S3 object key)
  └─ Returns signed download URL

POST /dev/update-content       → UpdateContent (dev only)
POST /dev/claude               → GetClaudeResponse (dev only)
```

## External Integrations

### D2L Integration

**Service:** `internal/services/d2l.go`

```go
type D2LClient struct {
  Cookies      map[string]string
  LocalStorage map[string]interface{}
  HTTPClient   *http.Client
}

// Fetches courses from D2L API
func (c *D2LClient) GetCoursesAndAssignments() ([]Course, error)

// Authenticates with saved cookies
func AuthenticateWithCookies(cookies map[string]string) (*D2LClient, error)
```

**Handlers:**
- `SaveCredentials` — Receives cookies/localStorage, stores encrypted
- `GetCoursesAndAssignments` — Retrieves user's D2L data
- `SyncCoursesAndAssignments` — Syncs to local database

### Claude AI Integration

**Service:** `internal/services/claude.go`

```go
// Direct Claude API call (non-agent)
func RunAssignmentCompletion(ctx context.Context, assignmentKey string, userID string) (string, error)

// Uses Anthropic SDK to send prompt to Claude, get response
```

**Anthropic SDK:** `github.com/anthropics/anthropic-sdk-go` (v1.30.0)
- Initialized from `ANTHROPIC_API_KEY` env var
- Used by both direct handlers and agent runner

### MinIO S3 Integration

**Service:** `internal/storage/s3.go`

```go
type BucketBasics struct {
  Client       *s3.Client
  BucketName   string
}

// Core operations
func (b *BucketBasics) PutObject(ctx context.Context, objectKey string, file []byte) error
func (b *BucketBasics) GetObject(ctx context.Context, objectKey string) ([]byte, error)
func (b *BucketBasics) ListObjects(ctx context.Context, prefix string) ([]string, error)
func (b *BucketBasics) PresignedURL(ctx context.Context, objectKey string) (string, error)
```

**AWS SDK:** `github.com/aws/aws-sdk-go-v2`
- Uses static credentials (MINIO_ROOT_USER, MINIO_ROOT_PASSWORD)
- Endpoint: MINIO_URL (internal) or MINIO_PUBLIC_URL (external via cloudflared)
- Bucket: Configured in env or defaults to "jaja"

## Environment Variables

**Root `.env` (docker-compose):**
```bash
POSTGRES_USER=jaja
POSTGRES_PASSWORD=secret
POSTGRES_DB=jaja
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

**`apps/server/.env` (Go server):**
```bash
PORT=8080                           # Server port
FRONTEND_URL=http://localhost:3000  # CORS origin
DB_URL=postgres://...               # PostgreSQL DSN
REDIS_URL=localhost:6379            # Redis endpoint (REQUIRED)
MINIO_URL=http://minio:9000         # MinIO internal URL
MINIO_PUBLIC_URL=...                # Cloudflared tunnel URL (presigned URLs)
AWS_REGION=us-east-1                # S3 region
MINIO_ROOT_USER=minioadmin          # S3 credentials
MINIO_ROOT_PASSWORD=minioadmin
ANTHROPIC_API_KEY=sk-...            # Claude API key (REQUIRED)
```

## Startup Flow

1. **Load .env** — `godotenv.Load()`
2. **Connect Database** — `database.ConnectDB()` → GORM + PostgreSQL
3. **Connect Redis** — `queue.ConnectRedis()` → asynq client
4. **Connect Storage** — `storage.ConnectObjectStorage()` → AWS SDK + MinIO
5. **Connect Agent** — `agent.ConnectAgent()` → ADK runner
6. **Start Workers** — `workers.Connect()` → asynq server starts polling jobs
7. **Setup Router** — Gin with CORS, register routes
8. **Listen** — `router.Run(":" + port)`

**Cleanup:** Deferred closes for Redis, worker shutdown

## Data Flow: Credential Storage

```
User Form (frontend)
  │ cookies + localStorage
  ▼
POST /d2l/credentials
  │
  ├─ Parse form data
  ├─ Validate JSON structure
  ├─ Create D2LCookieSession + D2LLocalStorageSession
  │
  ▼
Database INSERT
  │
  └─ Session stored with user_id + expiry
```

## Data Flow: Course Sync

```
User clicks "Sync Courses"
  │
  ▼
POST /d2l/sync
  │
  ├─ Retrieve D2LCookieSession from DB
  ├─ Create D2LClient with stored cookies
  │
  ▼
D2LClient.GetCoursesAndAssignments()
  │
  ├─ HTTP requests to D2L API
  ├─ Parse JSON responses
  │
  ▼
Database INSERT (courses, assignments)
  │
  └─ Frontend updates with course list
```

## Related Areas

- **Agent & Workers:** Job processing triggered by handlers
- **Frontend:** Sends credentials, triggers syncs
- **Migrations:** Schema evolution tracked in Goose files

## Dependencies

**Database:**
- `gorm.io/gorm` — ORM
- `gorm.io/driver/postgres` — PostgreSQL driver
- `github.com/lib/pq` — Pure Go PostgreSQL driver

**Queue:**
- `github.com/hibiken/asynq` — Job queue
- `github.com/redis/go-redis/v9` — Redis client

**Storage:**
- `github.com/aws/aws-sdk-go-v2` — AWS SDK
- `github.com/aws/aws-sdk-go-v2/service/s3` — S3 service

**HTTP:**
- `github.com/gin-gonic/gin` — Web framework
- `github.com/gin-contrib/cors` — CORS middleware

**AI:**
- `github.com/anthropics/anthropic-sdk-go` — Claude API

## Testing

### Integration Tests

Test database operations:
```bash
# Requires PostgreSQL + Docker
go test ./internal/models/...
go test ./internal/services/...
```

### Manual Testing

```bash
# Health check
curl http://localhost:8080/health

# Save credentials
curl -X POST http://localhost:8080/d2l/credentials \
  -F "cookies=..." \
  -F "localStorage=..."

# Fetch courses
curl http://localhost:8080/d2l/courses

# Upload assignment file
curl -X POST http://localhost:8080/dev/assignment-files \
  -F "file=@assignment.pdf"
```

## Future Improvements

1. **Authentication:** Replace hardcoded test user with real auth (OAuth2, JWT)
2. **Encryption:** Encrypt stored D2L cookies and localStorage at rest
3. **Connection Pooling:** Optimize database pool settings per load
4. **Rate Limiting:** Add rate limits to API endpoints
5. **Logging:** Structured logging with correlation IDs for tracing
6. **Metrics:** Prometheus metrics for job success/failure rates
7. **Caching:** Cache D2L course list with TTL
8. **Error Handling:** Standardized error responses with codes

# jaja

JAJA: Just Automate Junk Assignments
A web app with a Next.js frontend and Go backend.

## Prerequisites

- [Bun](https://bun.sh)
- [Go](https://go.dev) 1.25+

## Setup

### Frontend

```bash
cd frontend
bun install
bun dev
```

Runs at `http://localhost:3000`

### Backend

```bash
cd backend
go mod download
go run .
```

### Database

Populate your `.env` in the project root directory

```
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=example_db
```

Run the docker container

```bash
docker compose up
```

## Tech Stack

- **Frontend**: Next.js, React, Tailwind CSS, TypeScript
- **Backend**: Go, Gin
- **Database** PostgreSQL

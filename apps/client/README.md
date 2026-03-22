# JAJA — Client

Next.js 16 (App Router) frontend for JAJA. Built with React 19, Tailwind CSS 4, and shadcn/ui.

## Setup

```bash
bun install
bunx skills add vercel-labs/next-browser  # AI agent debugging tools (installs Chromium via Playwright)
```

Copy the environment file and configure:

```bash
cp .env.example .env  # set NEXT_PUBLIC_API_URL (default: http://localhost:8080)
```

## Development

```bash
bun dev       # Dev server on http://localhost:3000
bun build     # Production build
bunx eslint . # Lint
```

Or run the full stack from the repo root with `docker compose up`.

## Project Structure

```
app/              # Pages and app-level components
  components/     # Page-specific components (e.g., cookie-form.tsx)
components/ui/    # shadcn UI primitives (button, checkbox, field, etc.)
lib/utils.ts      # cn() helper (clsx + tailwind-merge)
utils/string.ts   # parseStringToJSON() for tab-separated cookie/storage data
```

## AI Agent Tools

This project includes `next-browser` for AI agent debugging. See [AGENTS.md](./AGENTS.md) for full usage docs and commands.

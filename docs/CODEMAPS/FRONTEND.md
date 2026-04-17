# Frontend Codemap

**Last Updated:** 2026-04-16

## Overview

The frontend is a Next.js 16 app with React 19 and Tailwind CSS 4, using shadcn/ui for components. It provides forms for D2L credential input, course synchronization, and assignment file uploads. Development utilities are in the `/dev/files` route for testing the agent.

## Architecture

```
┌──────────────────────────────────────────────────────┐
│           Next.js App Router (React 19)               │
└──────────┬───────────────────────────────────────────┘
           │
    ┌──────┴──────┬──────────┬──────────────┐
    │             │          │              │
    ▼             ▼          ▼              ▼
┌────────┐  ┌─────────┐  ┌────────┐  ┌──────────────┐
│ Pages  │  │Component│  │ Utils  │  │ shadcn/ui    │
│ /      │  │s        │  │ (lib)  │  │ (Radix + TW) │
│ /dev   │  │ (12+)   │  │        │  └──────────────┘
└────────┘  └─────────┘  └────────┘
    │             │          │
    └─────────────┴──────────┘
           │
           ▼
    ┌──────────────────────┐
    │  Tailwind CSS 4      │
    │  + Radix Lyra style  │
    │  + Phosphor Icons    │
    └──────────────────────┘
           │
           ▼
    ┌──────────────────────┐
    │  API Calls           │
    │  http://localhost:.. │
    │  /d2l/*              │
    │  /dev/*              │
    └──────────────────────┘
```

## Directory Structure

```
apps/client/
├── app/                      # Next.js App Router
│   ├── layout.tsx            # Root layout, global styles
│   ├── page.tsx              # Home page (index)
│   ├── dev/
│   │   ├── layout.tsx        # Dev section layout
│   │   ├── page.tsx          # Dev dashboard
│   │   ├── files/
│   │   │   ├── page.tsx      # Assignment file upload
│   │   │   └── components/   # File-related components
│   │   └── components/       # Dev-specific components
│   │       ├── cookie-form.tsx
│   │       ├── sync-courses.tsx
│   │       └── ...
│   └── components/           # Page-level components
├── components/               # Reusable UI components
│   ├── ui/                   # shadcn/ui primitives
│   │   ├── button.tsx
│   │   ├── input.tsx
│   │   ├── form.tsx
│   │   ├── checkbox.tsx
│   │   └── ... (20+ primitives)
│   └── ...                   # Custom components
├── lib/
│   └── utils.ts              # cn() helper, Tailwind merge
├── utils/
│   └── string.ts             # parseStringToJSON()
├── public/                   # Static assets
├── components.json           # shadcn/ui config (Radix Lyra)
├── tsconfig.json             # TypeScript config (@/* alias)
├── next.config.js            # Next.js config
├── tailwind.config.ts        # Tailwind CSS 4 config
└── package.json              # Dependencies (React 19, Tailwind 4, etc.)
```

## Key Modules

| Component | Purpose | Location | Responsibility |
|-----------|---------|----------|-----------------|
| **Root Layout** | Global styles, fonts | `app/layout.tsx` | Wraps all pages, applies Tailwind reset, fonts |
| **Home Page** | Landing page | `app/page.tsx` | Intro, links to /dev section |
| **Dev Dashboard** | Dev utilities hub | `app/dev/page.tsx` | Overview of testing endpoints |
| **Dev File Upload** | Assignment upload UI | `app/dev/files/page.tsx` | File picker, upload form, S3 integration |
| **Cookie Form** | D2L credential input | `app/dev/components/cookie-form.tsx` | Textarea for cookies, localStorage JSON |
| **Sync Courses** | Course sync button | `app/dev/components/sync-courses.tsx` | Fetches courses from D2L, displays list |
| **shadcn/ui Primitives** | Radix UI + Tailwind | `components/ui/*.tsx` | Button, Input, Checkbox, Dialog, Form, Textarea, etc. |
| **Utils** | Helper functions | `lib/utils.ts`, `utils/string.ts` | `cn()` for Tailwind merge, `parseStringToJSON()` |

## Pages

### Home (`app/page.tsx`)

**Route:** `/`

**Components:**
- Hero section with project description
- Call-to-action links to `/dev`
- Tech stack badges

**API Calls:** None (static)

### Dev Dashboard (`app/dev/page.tsx`)

**Route:** `/dev`

**Components:**
- Overview of available dev endpoints
- Links to file upload, test panels
- Documentation snippets

**API Calls:** None (static)

### Dev File Upload (`app/dev/files/page.tsx`)

**Route:** `/dev/files`

**Components:**
- File input (PDF, images)
- Submit button
- Upload progress indicator
- Response display

**API Flow:**
```
User selects file
    │
    ▼
Validate file type/size
    │
    ▼
POST /dev/assignment-files (multipart form)
    │
    ├─ Uploads to MinIO S3
    ├─ Server generates S3 key
    │
    ▼
Display success + S3 URL
    │
    └─ User can download/reference file
```

**Success Response:**
```json
{
  "url": "http://minio:9000/jaja/assignment-123.pdf",
  "key": "assignment-123.pdf"
}
```

## Components

### Dev Components (`app/dev/components/`)

#### cookie-form.tsx
- Textarea for D2L cookies (JSON or serialized)
- Textarea for localStorage (JSON)
- Submit button
- API: `POST /d2l/credentials`
- Stores credentials server-side

#### sync-courses.tsx
- Fetch button
- Displays list of synced courses
- Shows assignment count per course
- API: `GET /d2l/courses`, `POST /d2l/sync`

### shadcn/ui Primitives

All components in `components/ui/` follow Radix UI patterns with Tailwind CSS:

- **button.tsx** — Styled button with variants (default, outline, ghost)
- **input.tsx** — Text input with focus states
- **form.tsx** — Form wrapper with context (react-hook-form compatible)
- **checkbox.tsx** — Checkbox with label
- **textarea.tsx** — Multiline text input
- **dialog.tsx** — Modal dialog (AlertDialog compatible)
- **select.tsx** — Dropdown select
- **label.tsx** — Form label
- **card.tsx** — Card container
- **badge.tsx** — Badge/tag component
- **toggle.tsx** — Toggle button
- **tabs.tsx** — Tabbed interface
- **dropdown-menu.tsx** — Dropdown menu
- **etc.** — 20+ primitives available

## Styling System

### Tailwind CSS 4

**Config:** `tailwind.config.ts`

**Features:**
- Modern v4 with content-based purging
- Responsive breakpoints (sm, md, lg, xl, 2xl)
- Dark mode support (class strategy)
- Extended color palette
- CSS Grid and Flexbox utilities

### shadcn/ui Radix Lyra Style

**Config:** `components.json`

```json
{
  "style": "new-york",
  "baseColor": "slate"
}
```

**Icons:** Phosphor Icons (via `phosphor-react`)

Example icon usage:
```tsx
import { Folder, Upload } from "phosphor-react"

<Upload size={24} />
```

### Utility Functions

**`lib/utils.ts`:**
```typescript
import { clsx, type ClassValue } from "clsx"
import { twMerge } from "tailwind-merge"

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs))
}
```

Used for conditional Tailwind class merging:
```tsx
<button className={cn("px-4 py-2", isActive && "bg-blue-500")}>
  Click me
</button>
```

## Utils

### String Utilities (`utils/string.ts`)

**`parseStringToJSON()`** — Converts tab-separated or JSON strings to objects

Used for parsing D2L cookies and localStorage from textarea input:

```typescript
export function parseStringToJSON(str: string): Record<string, any> {
  try {
    // Try JSON parse first
    return JSON.parse(str)
  } catch (e) {
    // Fallback: parse tab-separated key=value pairs
    const obj: Record<string, any> = {}
    str.split("\n").forEach((line) => {
      const [key, value] = line.split("=")
      if (key && value) obj[key.trim()] = value.trim()
    })
    return obj
  }
}
```

## API Integration

### Endpoints Called

**D2L Endpoints:**
```typescript
// Save credentials
fetch("/d2l/credentials", {
  method: "POST",
  body: formData,  // FormData with cookies, localStorage
})

// Get courses
fetch("/d2l/courses")

// Sync courses
fetch("/d2l/sync", {
  method: "POST",
  body: JSON.stringify({ course_ids, assignment_ids }),
})
```

**Dev Endpoints:**
```typescript
// Upload files
fetch("/dev/assignment-files", {
  method: "POST",
  body: formData,  // FormData with files
})

// Run agent
fetch("/dev/run-agent", {
  method: "POST",
  body: JSON.stringify({
    session_id: "user-session",
    prompt: "Complete this essay...",
    user_id: "uuid",
  }),
})

// Generate presigned URL
fetch("/dev/presigned-url?path=result.docx")
```

## Type Safety

**TypeScript enabled** with path aliases:

```json
{
  "compilerOptions": {
    "baseUrl": ".",
    "paths": {
      "@/*": ["./*"]
    }
  }
}
```

Example import:
```typescript
import { Button } from "@/components/ui/button"
import { cn } from "@/lib/utils"
```

## Dependencies

**package.json (key libraries):**
- `react@19` — Latest React
- `next@16` — Next.js with App Router
- `tailwindcss@4` — Latest Tailwind
- `radix-ui/*` — Unstyled component primitives
- `phosphor-react` — Icon library
- `react-hook-form` — Form state management
- `zod` — Schema validation
- `clsx` + `tailwind-merge` — Utility functions

**Dev Dependencies:**
- `@types/react`, `@types/node` — TypeScript definitions
- `typescript` — TypeScript compiler
- `autoprefixer`, `postcss` — CSS processing
- `eslint` — Linting

## Build & Deployment

### Development

```bash
cd apps/client
bun install
bun dev
```

Runs at http://localhost:3000 with hot reload

### Production Build

```bash
bun build      # Generates .next/ directory
bun start      # Runs production server
```

### Docker

The frontend runs in Docker via docker-compose:
```dockerfile
FROM node:20-alpine
WORKDIR /app
COPY . .
RUN bun install && bun build
EXPOSE 3000
CMD ["bun", "start"]
```

Exposed as http://localhost:3000 in compose

## Performance Optimizations

1. **Code Splitting:** Next.js automatically splits routes
2. **Image Optimization:** Built-in next/image component
3. **CSS Minification:** Tailwind purges unused styles
4. **Font Optimization:** Google Fonts via next/font
5. **API Route Caching:** SWR for data fetching (optional)

## Accessibility

- shadcn/ui components use Radix primitives (accessible by default)
- Keyboard navigation on all interactive elements
- ARIA labels on form inputs
- Focus indicators on buttons and links

## Testing

### Manual Testing

```bash
bun dev
# Open http://localhost:3000
# Test: Home → /dev → /dev/files
# Upload file, see success response
```

### Component Testing (Future)

```bash
# vitest + @testing-library/react
bun test
```

### E2E Testing (Future)

```bash
# playwright + @vercel/next-browser
bun exec -- bunx next-browser test
```

## Future Enhancements

1. **Real-time Job Status:** WebSocket updates for agent progress
2. **Job History:** Display past completion jobs with results
3. **Multi-file Upload:** Handle multiple assignments at once
4. **Editor Integration:** Show assignment completion in real-time
5. **Dark Mode Toggle:** Switch between light/dark themes
6. **Mobile Responsiveness:** Optimize for tablets and phones
7. **Error Boundaries:** Graceful error handling with recovery
8. **Offline Support:** Service Worker caching

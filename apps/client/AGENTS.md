<!-- BEGIN:nextjs-agent-rules -->

# Next.js: ALWAYS read docs before coding

Before any Next.js work, find and read the relevant doc in `node_modules/next/dist/docs/`. Your training data is outdated — the docs are the source of truth.

<!-- END:nextjs-agent-rules -->

## AI Agent Tools: next-browser

This project includes `next-browser` — a CLI that gives Claude Code and other agents access to React DevTools data like component trees, props, hooks, errors, and performance metrics as shell commands.

### Installation

The next-browser skill is installed via:
```bash
bunx skills add vercel-labs/next-browser
```

During installation, Playwright will prompt to install Chromium. Approve it. This sets up:
- `@vercel/next-browser` CLI binary
- Chromium browser
- Ability to run headed or headless (`NEXT_BROWSER_HEADLESS=1`)

The skill is stored in `.agents/skills/next-browser/` and version is tracked in `skills-lock.json`.

### Usage with Claude Code

When using Claude Code on this project, start with `/next-browser` to invoke the skill. The browser opens with access to:
- **Component tree** — Full React hierarchy, IDs, keys
- **Props & state** — Inspect any component's props, hooks, state
- **Errors** — Build and runtime errors for the current page
- **Partial Pre-Rendering** — Debug PPR shells with `ppr lock`/`ppr unlock`
- **Network** — Inspect requests and responses
- **Accessibility** — Snapshot with interactive element markers
- **Screenshots** — Visual previews with captions

### Common Commands

```bash
next-browser open http://localhost:3000           # Launch browser
next-browser snapshot                              # Accessibility tree
next-browser tree                                  # Full component tree
next-browser tree <component-id>                  # Inspect one component
next-browser click e0                              # Click element by ref
next-browser fill <selector> <value>              # Fill input
next-browser preview "Caption"                    # Screenshot
next-browser ppr lock                             # Freeze dynamic content (PPR)
next-browser reload                               # Reload page
next-browser close                                # Close browser & daemon
```

See `.agents/skills/next-browser/README.md` for full command reference.

### Daemon Cache Gotcha

The daemon keeps running the old build in memory. After `bun build`, you must restart the daemon:
```bash
next-browser close      # Kill daemon
# Then reopen with: next-browser open ...
```

Without this, you'll be testing stale code.

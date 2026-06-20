# AGENTS.md

This file provides context for AI agents working on the Portkey codebase.

---

## Project Overview

**Portkey** is a lightweight Go web portal / start page that aggregates bookmarked links. It supports full-text and fuzzy search, custom HTML pages, dark/light mode, Prometheus metrics, and live-reload development. There is no database — all content comes from a single YAML config file.

- **Language:** Go 1.26 (`CGO_ENABLED=0`, fully static binaries)
- **Frontend:** [Templ](https://github.com/a-h/templ) templates, Tailwind CSS v4, Alpine.js, HTMX
- **Config:** `config.yml` (YAML) with `PORTKEY_*` environment variable overrides
- **License:** AGPL-3.0

---

## Repository Layout

```
portkey/
├── main.go                   # Entry point: loads config, starts HTTP server
├── main_test.go              # Integration test (startup probe via /healthz)
├── config.yml                # Example configuration file
├── go.mod / go.sum           # Go modules (declared dependencies)
├── package.json              # Frontend tooling (Tailwind, esbuild, Alpine, HTMX)
├── build.sh                  # Binary build script (injects version/commit via ldflags)
├── build-docker.sh           # Docker image build wrapper
├── Dockerfile                # Multi-stage Docker build (node → go → alpine)
├── fly.toml                  # Fly.io deployment configuration
├── mise.toml                 # Tool version management
├── cliff.toml                # Changelog generation (conventional commits)
├── assets/
│   ├── css/main.css          # Tailwind CSS source (input)
│   └── js/main.js            # Alpine.js + HTMX source (input)
├── static/
│   ├── css/main.css          # Compiled Tailwind CSS (output — do not edit manually)
│   └── js/main.js            # Bundled JS via esbuild (output — do not edit manually)
├── docs/                     # Documentation assets and screenshots
└── internal/
    ├── build/                # Compile-time metadata (Version, CommitHash, BuildTime, GoVersion)
    ├── components/           # Templ HTML templates (*_templ.go are auto-generated)
    ├── config/               # Configuration loading (Viper + flags)
    ├── favicon/              # Disk-backed favicon cache (fetch, persist, serve)
    ├── metrics/              # Prometheus metrics definitions
    ├── models/               # Core data structures (Portal, Page)
    ├── server/               # HTTP mux, routes, and all handlers
    └── utils/                # Shared helpers (middleware, test utilities, array helpers)
```

---

## Build & Run

### Go Binary
```bash
# Build with version injection
./build.sh [-v <version>] [-o <output-path>]

# Run (reads config.yml from current directory by default)
./portkey

# Custom config directory
./portkey --config-path=/etc/portkey/
```

### Frontend Assets
```bash
npm install --include=dev   # Install frontend tooling
npm run build               # Compile Tailwind CSS + bundle JS
npm run watch               # Watch mode for development
```

### Live Reload (development)
```bash
go tool air                 # Uses .air.toml; rebuilds on file changes
```

### Docker
```bash
./build-docker.sh           # Builds image tagged codehat/portkey:local

# Run with mounted config
docker run -v $(PWD)/config.yml:/opt/config.yml -p 3000:3000 codehat/portkey:latest
```

---

## Testing

```bash
go test ./...
```

The main integration test (`main_test.go`) starts the server and polls `GET /healthz` (5-second timeout) to verify startup. There are no unit test mocks; prefer real handler/server tests.

---

## Key Design Patterns

### HTTP Handlers
All handlers are struct-based with a `handle()` method returning `http.HandlerFunc`. They receive a `*slog.Logger` at construction time.

```go
type searchHandler struct {
    logger *slog.Logger
}

func (s searchHandler) handle() http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) { ... }
}
```

### Global Singletons (loaded at startup)
| `config.C` | `internal/config` | Parsed application config |
| `config.F` | `internal/config` | CLI flags |
| `metrics.M` | `internal/metrics` | Prometheus metric counters |
| `build.B` | `internal/build` | Binary build metadata |
| `favicon.C` | `internal/favicon` | Favicon cache instance |

### Templ Templates
HTML is rendered using type-safe Templ components. **Never edit `*_templ.go` files directly** — they are auto-generated from `*.templ` sources.

```bash
go generate ./...   # Re-generates *_templ.go from *.templ
```

### Configuration
Config is loaded once at startup via `config.Load()`. All fields can be overridden with `PORTKEY_<FIELDNAME>` environment variables (Viper handles this).

```go
config.C.Title           // UI title
config.C.Portals         // []models.Portal
config.C.Pages           // []models.Page
config.C.EnableMetrics   // bool
config.C.LayoutColumns   // int (0 = vertical, 2-12 = column grid)
config.C.FaviconCacheDir // string (on-disk favicon cache path)
config.C.FaviconCacheDisabled // bool (bypass local cache)
```

---

## HTTP Routes

| Route | Method | Description |
|-------|--------|-------------|
| `/` | GET | Home page — renders portals grouped by their `group` field |
| `/_/favicon?domain=<hostname>` | GET | Serves cached favicon or fetches on miss |
| `/_/portals?search=<q>` | GET | HTMX partial — grouped view with optional search filtering |
| `/api/portals` | GET | JSON list of all portals |
| `/api/pages` | GET | JSON list of all pages |
| `/static/*` | GET | Embedded static assets (CSS, JS) |
| `/version` | GET | Build metadata page |
| `/healthz` | GET | Health check — returns `"OK"` |
| `/reload` | GET | Dev-mode browser reload (WebSocket) |
| `/metrics` | GET | Prometheus metrics (separate server, port 3030) |
| `/<portal-title>` | GET | Redirect to the portal's URL |
| `/<page-path>` | GET | Render a custom HTML page |

---

## Data Models

```go
// internal/models/models.go

type Portal struct {
    Link     string   // Absolute URL or relative path
    Title    string
    Emoji    string
    Keywords []string // Used in search
    Group    string   // Optional section heading on the home page (empty = ungrouped)
}

// Portal.IsExternal() bool        — true if Link starts with http
// Portal.TitleForUrl() string     — URL-safe version of Title

// PortalGroup groups portals under a shared name for home-page rendering.
type PortalGroup struct {
    Name    string
    Portals []Portal
}

type Page struct {
    Heading  string
    Subtitle string
    Content  string // Raw HTML allowed
    Path     string // Route path, e.g. "/about"
}
```

---

## Configuration Reference

```yaml
# Server
logLevel: INFO         # ERROR | WARN | INFO | DEBUG
logJson: false
host: localhost
port: 3000
contextPath: ""        # Mount under a subpath, e.g. /portkey
enableMetrics: true
metricsHost: localhost
metricsPort: 3030

# UI
title: "portkey"
hideTitle: false
subtitle: "Where do you want to go?"
hideSearchBar: false
showTopIcon: true
showKeywordsAsTooltips: false
sortAlphabetically: false
headerAddition: ""     # Injected into <head>
footer: ""             # Injected into page footer

# Search
searchWithStringSimilarity: true
minimumStringSimilarity: 0.5   # 0.0–1.0; Levenshtein threshold

# Layout
layoutColumns: 0             # 0 = vertical (default), 2-12 = multi-column grid. On mobile (<768px) always vertical.

# Favicon
faviconServiceURL: https://favicon.vemetric.com  # Self-hostable favicon fetch service
faviconCacheDir: ./favicon-cache                # On-disk favicon cache (Docker volume mountable)
faviconCacheDisabled: false                     # Bypass local cache, fetch directly from remote

# Portals
portals:
  - title: "GitHub"
    emoji: 💻
    link: https://github.com/
    keywords: [code, git]
    group: Development   # optional — organises portal under a named section heading

# Custom pages
pages:
  - heading: "About"
    subtitle: "Info"
    path: /about
    content: "<p>Custom HTML</p>"
```

---

## Prometheus Metrics

All metrics are registered under the `portkey_` namespace.

| Metric | Type | Labels |
|--------|------|--------|
| `portkey_portal_handler_requests_total` | Counter | `portal` |
| `portkey_page_handler_requests_total` | Counter | `path` |
| `portkey_search_requests_with_results_total` | Counter | — |
| `portkey_search_requests_no_results_total` | Counter | — |
| `portkey_search_duration_seconds` | Histogram | — |
| `portkey_http_request_duration_seconds` | Histogram | `handler` |
| `portkey_favicon_cache_hits_total` | Counter | — |
| `portkey_favicon_cache_misses_total` | Counter | — |
| `portkey_favicon_fetch_failures_total` | Counter | — |
| `portkey_favicon_cache_size` | Gauge | — |
| `portkey_portals_total` | Gauge | — |
| `portkey_groups_total` | Gauge | — |
| `portkey_version_info` | Gauge | `version`, `buildTime`, `commitHash`, `goVersion` |

---

## CI/CD (GitHub Actions)

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| `tests.yml` | push / PR | `go test ./...` |
| `release-build.yml` | release | Cross-compile binaries |
| `build-docker-latest.yml` | push to main | Build & push `latest` Docker tag |
| `build-docker-release.yml` | release | Build & push version-tagged Docker image |
| `fly.yml` | push to main | Deploy to Fly.io |
| `update-changelog-file.yml` | release | Generate CHANGELOG.md via git-cliff |

---

## Important Notes for Agents

- **Do not edit `*_templ.go` files.** Edit the corresponding `*.templ` file and run `go generate ./...`.
- **Do not edit `static/css/main.css` or `static/js/main.js` directly.** Edit sources in `assets/` and run `npm run build`.
- The CSS file hash in `internal/build/` is computed from the content of `static/css/main.css` for cache-busting — it updates automatically at build time.
- `config.C`, `metrics.M`, and `build.B` are globals initialized at startup; access them directly from handlers rather than passing through the call stack.

- **Favicon caching:** `internal/favicon/cache.go` provides a disk-backed favicon cache. Favicons are fetched from a configurable remote service (`faviconServiceURL`, default `https://favicon.vemetric.com`), stored by normalized hostname in `faviconCacheDir` (default `./favicon-cache`), and served via `GET /_/favicon?domain=<hostname>`. Cache TTL is 7 days; stale entries are refreshed in the background. Failed fetches are backed off for 1 hour. Set `faviconCacheDisabled: true` to bypass local caching entirely.
- **Multi-column layout:** Set `config.C.LayoutColumns` to 0 (vertical, default) or 2-12 (CSS grid). The `components.GridClass(columns int)` helper returns responsive Tailwind classes: mobile uses `max-md:flex flex-col items-start` (vertical stack), desktop uses `md:grid md:grid-cols-N`. Class strings are stored in a `[...]string` array with literal entries so Tailwind v4 detects them during CSS build; always add new column values as literal strings in `internal/components/component.go`.
- The server uses the standard library `net/http` mux — no third-party router.
- All binaries are statically linked (`CGO_ENABLED=0`); avoid importing packages that require CGO.
- Conventional Commits are used; `cliff.toml` drives changelog generation.
- **Portal grouping:** `internal/utils/portal.go` provides `GroupPortals(portals []models.Portal) []models.PortalGroup`. Named groups appear in definition order; portals with an empty `Group` field are collected into an unnamed group placed last. The search handler (`/_/portals`) returns groups even when a search query is active (filtered portals are grouped). `GroupedPortalPartial(groups, columns int)` renders groups in a grid when `columns > 0`. `PortalPartial(portals, columns int)` handles the non-grouped case.

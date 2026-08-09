# Breaking Changes

Portkey follows [Semantic Versioning](https://semver.org/spec/v2.0.0.html): breaking changes are
only introduced in **major** version releases (e.g. `4.x.y`). Minor and patch
releases never break the configuration, environment variables, routes, or the CLI.

This document lists, for each major version, the changes that require end users
(people deploying and configuring Portkey) to update their setup.

## Index

- [3.x.y → 4.x.y](#v4)
<!-- When a new major version ships, add its entry to the index and a new section below. -->

---

<a id="v4"></a>
## 3.x.y → 4.x.y

### Portal field `emoji` renamed to `icon`

**What changed**

The `emoji` field on portal entries in `config.yml` has been renamed to `icon`.

| Before (3.x.y) | After (4.x.y) |
|---|---|
| `emoji: 💻` | `icon: 💻` |

The `icon` field now accepts more than just emojis:

- An **emoji** (same as the old `emoji` field), e.g. `icon: 💻`
- A **relative path** to an image file, e.g. `icon: /static/logo.svg` (or a file
  served from a `favicon.customIconsDir` directory)
- An **absolute URL** to an image, e.g. `icon: https://example.com/logo.png`

When `icon` is empty/omitted, the behaviour is now:

- **External links** → the site's favicon is fetched and displayed automatically.
- **Internal links** → a default file/document icon is shown.

**Why**

Portal icons were generalised to support images and automatic favicons as part of the
4.x.y portal redesign.

**Required action**

Rename every `emoji:` key to `icon:` in your `config.yml`. Existing `emoji:` entries
are **silently ignored** after the upgrade (the field no longer exists), so portals
would lose their emoji and external links would fall back to auto-fetched favicons.

```yaml
# Before
portals:
  - title: GitHub
    emoji: 💻
    link: https://github.com/

# After
portals:
  - title: GitHub
    icon: 💻
    link: https://github.com/
```

---

### Favicon cache directory must be writable at startup

**What changed**

4.x.y introduces on-disk favicon caching, **enabled by default**. The
`favicon.cacheDir` option defaults to `./favicon-cache` (relative to the working
directory). At startup Portkey unconditionally creates this directory; if it cannot,
**Portkey panics and exits immediately**.

This affects any deployment where the Portkey process cannot write to its working
directory, most notably **Docker**: the official image runs as the non-root user
`nonroot` with `WORKDIR /opt`, but `/opt` is owned by `root` (`drwxr-xr-x`), so
`nonroot` cannot create `/opt/favicon-cache`. Upgrading such a deployment to 4.x.y
makes the container crash-loop on startup.

**Required action**

Point `favicon.cacheDir` at a path the Portkey process can write to.

- **Docker** — mount a writable volume and direct the cache there, either via config:

  ```yaml
  favicon:
    cacheDir: /data/favicon-cache
  ```

  or via the environment variable:

  ```sh
  docker run -v portkey-cache:/data \
    -e PORTKEY_FAVICON_CACHE_DIR=/data/favicon-cache \
    codehat/portkey:latest
  ```

- **Binary / bare-metal** — if you run Portkey as a non-privileged user from a
  directory that user does not own (e.g. installed under a root-owned path), set
  `favicon.cacheDir` (or `PORTKEY_FAVICON_CACHE_DIR`) to a writable location such as
  `/tmp/favicon-cache` or a dedicated data directory. If you run Portkey from a
  directory you own, no change is needed.

> **Note:** Setting `favicon.cacheEnabled: false` does **not** prevent this startup
> panic — the directory is created regardless. A writable `favicon.cacheDir` is always
> required.

**Related new behaviour**

For external portals without an explicit `icon`, Portkey now makes outbound HTTP
requests to fetch and cache the site's favicon. In air-gapped or network-restricted
environments, either set an `icon` per portal or disable fetching with
`favicon.cacheEnabled: false` (`PORTKEY_FAVICON_CACHEENABLED=false`) — while still
providing a writable `favicon.cacheDir` as described above.

---

### Configuration restructured into groups

**What changed**

All configuration options are now grouped under `server`, `metrics`, `ui`, `search`,
and `favicon` sections. Options were renamed to positive forms (`hide*` → `show*`,
`enable*` → `enabled`, `*Disabled` → `*Enabled`) and environment variables now use
underscores for the group prefix (`PORTKEY_SERVER_HOST` instead of `PORTKEY_HOST`).

| Before (3.x.y) | After (4.x.y) | Env var after |
|---|---|---|
| `logLevel`, `logJson`, `host`, `port`, `contextPath`, `devMode` | `server.logLevel`, `server.logJson`, `server.host`, `server.port`, `server.contextPath`, `server.devMode` | `PORTKEY_SERVER_LOGLEVEL`, `PORTKEY_SERVER_PORT`, … |
| `enableMetrics`, `metricsHost`, `metricsPort` | `metrics.enabled`, `metrics.host`, `metrics.port` | `PORTKEY_METRICS_ENABLED`, `PORTKEY_METRICS_HOST`, `PORTKEY_METRICS_PORT` |
| `title`, `hideTitle`, `hideSearchBar` | `ui.title`, `ui.showSearchBar` (inverted). `ui.showTitle` was **removed** — an empty `ui.title: ""` hides the top-bar title now | `PORTKEY_UI_TITLE`, `PORTKEY_UI_SHOWSEARCHBAR` |
| `showTopIcon`, `showKeywordsAsTooltips`, `sortAlphabetically`, `layoutColumns`, `headerAddition`, `footer` | `ui.showTopIcon`, `ui.showKeywordsAsTooltips`, `ui.sortAlphabetically`, `ui.layoutColumns`, `ui.headerAddition`, `ui.footer` | `PORTKEY_UI_SHOWTOPICON`, `PORTKEY_UI_LAYOUTCOLUMNS`, … |
| `searchWithStringSimilarity`, `minimumStringSimilarity` | `search.stringSimilarity`, `search.minimumSimilarity` | `PORTKEY_SEARCH_STRINGSIMILARITY`, `PORTKEY_SEARCH_MINIMUMSIMILARITY` |
| `faviconMode`, `faviconServiceURL`, `faviconCacheDir`, `faviconCacheDisabled`, `customIconsDir` | `favicon.mode`, `favicon.serviceUrl`, `favicon.cacheDir`, `favicon.cacheEnabled` (inverted), `favicon.customIconsDir` | `PORTKEY_FAVICON_MODE`, `PORTKEY_FAVICON_SERVICEURL`, `PORTKEY_FAVICON_CACHE_DIR`, `PORTKEY_FAVICON_CACHEENABLED`, `PORTKEY_FAVICON_CUSTOMICONSDIR` |

The boolean flags changed meaning:

| Before (3.x.y) | After (4.x.y) |
|---|---|
| `hideTitle: true` | `ui.title: ""` (option `ui.showTitle` was removed) |
| `hideSearchBar: true` | `ui.showSearchBar: false` |
| `faviconCacheDisabled: true` | `favicon.cacheEnabled: false` |

**Why**

The flat option list grew unwieldy and mixed negative (`hide*`) and positive
(`show*`) forms. Grouping mirrors the config file structure and positive forms are
less error-prone (a zero value no longer accidentally hides UI elements).

**Required action**

Migrate `config.yml` to the grouped layout and update `PORTKEY_*` environment
variables to the new names. Unknown keys (e.g. leftover flat options) now **fail
startup** with a decoding error instead of being silently ignored.

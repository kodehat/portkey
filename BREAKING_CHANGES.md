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

Rename every `emoji:` key to `icon:` in your `config.yml`. Unknown keys now **fail
startup** with a decoding error, so leftover `emoji:` entries must be renamed —
they are no longer silently ignored.

The same rename applies to the JSON API: `/api/portals` now returns `icon` instead
of `emoji` for each portal.

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

### Configuration restructured into groups

**What changed**

All configuration options are now grouped under `server`, `metrics`, `ui`, `search`,
and `favicon` sections. Options were renamed to positive forms (`hide*` → `show*`,
`enable*` → `enabled`) and environment variables now use underscores for the group
prefix (`PORTKEY_SERVER_HOST` instead of `PORTKEY_HOST`).

| Before (3.x.y) | After (4.x.y) | Env var after |
|---|---|---|
| `logLevel`, `logJson`, `host`, `port`, `contextPath`, `devMode` | `server.logLevel`, `server.logJson`, `server.host`, `server.port`, `server.contextPath`, `server.devMode` | `PORTKEY_SERVER_LOGLEVEL`, `PORTKEY_SERVER_PORT`, … |
| `enableMetrics`, `metricsHost`, `metricsPort` | `metrics.enabled`, `metrics.host`, `metrics.port` | `PORTKEY_METRICS_ENABLED`, `PORTKEY_METRICS_HOST`, `PORTKEY_METRICS_PORT` |
| `title`, `hideTitle`, `hideSearchBar` | `ui.title`, `ui.showSearchBar` (inverted). `ui.showTitle` was **removed** — an empty `ui.title: ""` hides the top-bar title now | `PORTKEY_UI_TITLE`, `PORTKEY_UI_SHOWSEARCHBAR` |
| `showTopIcon`, `showKeywordsAsTooltips`, `sortAlphabetically`, `headerAddition`, `footer` | `ui.showTopIcon`, `ui.showKeywordsAsTooltips`, `ui.sortAlphabetically`, `ui.headerAddition`, `ui.footer` | `PORTKEY_UI_SHOWTOPICON`, `PORTKEY_UI_HEADERADDITION`, … |
| `searchWithStringSimilarity`, `minimumStringSimilarity` | `search.stringSimilarity`, `search.minimumSimilarity` | `PORTKEY_SEARCH_STRINGSIMILARITY`, `PORTKEY_SEARCH_MINIMUMSIMILARITY` |

The following options are **new in 4.x.y** (they did not exist in 3.x.y, so there is
nothing to migrate): `layoutColumns` (now `ui.layoutColumns`), and the whole
`favicon` group (`favicon.mode`, `favicon.serviceUrl`, `favicon.cacheDir`,
`favicon.cacheEnabled`, `favicon.customIconsDir`).

The boolean flags changed meaning:

| Before (3.x.y) | After (4.x.y) |
|---|---|
| `hideTitle: true` | `ui.title: ""` (option `ui.showTitle` was removed) |
| `hideSearchBar: true` | `ui.showSearchBar: false` |

**Why**

The flat option list grew unwieldy and mixed negative (`hide*`) and positive
(`show*`) forms. Grouping mirrors the config file structure and positive forms are
less error-prone (a zero value no longer accidentally hides UI elements).

**Required action**

Migrate `config.yml` to the grouped layout and update `PORTKEY_*` environment
variables to the new names. Unknown keys (e.g. leftover flat options) now **fail
startup** with a decoding error instead of being silently ignored.

---

### Option `subtitle` removed

**What changed**

The global `subtitle` option (shown below the title in 3.x.y) has been removed in
4.x.y. The field no longer exists, and a `subtitle:` key in `config.yml` now **fails
startup** with a decoding error.

**Required action**

Remove the `subtitle:` key from your `config.yml`. There is no replacement — the
title area is now controlled solely by `ui.title` (see the configuration restructure
above).

# favicon-api

This directory holds the Fly.io configuration for the favicon service that
powers the portkey demo instance.

## What is deployed here?

The Fly.io app **`portkey-favicon-api`** runs the
[Vemetric favicon API](https://hub.docker.com/r/vemetric/favicon-api) Docker
image (`docker.io/vemetric/favicon-api:1.0.0`). Given a hostname, the service
discovers and returns that site's favicon. It is reachable at
`https://favicon-api.portkey.page`.

## Why?

The portkey demo (`https://demo.portkey.page`, Fly.io app `portkey-demo`) uses
this service as its last-resort favicon fallback via the `faviconServiceURL`
config key:

```yaml
faviconServiceURL: https://favicon-api.portkey.page
```

When the demo can't discover a favicon directly from a target website, it falls
back to this service.
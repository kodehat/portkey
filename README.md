<p align="center"><img src="docs/images/logo.png" alt="portkey logo"></p>

<h1 align="center">portkey</h1>
<div align="center">
  ⚡🚀🔗
</div>
<div align="center">
  <strong>Repository of the portkey application</strong>
</div>
<div align="center">
  A simple web portal that can act as startup page and shows a collection of links/urls. It also supports adding small custom pages.
</div>

<br />

<div align="center">
  <!-- Sonar quality gate -->
  <a href="https://sonarcloud.io/project/overview?id=kodehat_portkey">
    <img src="https://sonarcloud.io/api/project_badges/measure?project=kodehat_portkey&metric=alert_status"
      alt="Sonar quality gate" />
  </a>
  <!-- GitHub stars -->
  <a href="https://github.com/kodehat/portkey">
    <img src="https://img.shields.io/github/stars/kodehat/portkey"
      alt="GitHub stars" />
  </a>
  <!-- Go version -->
  <img src="https://img.shields.io/github/go-mod/go-version/kodehat/portkey"
      alt="Go version" />
  <!-- Docker image size -->
  <a href="https://hub.docker.com/r/codehat/portkey">
    <img src="https://img.shields.io/docker/image-size/codehat/portkey"
      alt="Docker image size" />
  </a>
</div>

<div align="center">
  <h3>
    <a href="https://www.portkey.page">
      Website (GitHub)
    </a>
    <span> | </span>
    <a href="https://github.com/kodehat/portkey/blob/main/.github/CONTRIBUTING.md">
      Contributing
    </a>
  </h3>
</div>

<div align="center">
  <sub>As magical as in Harry Potter. Built with ☕️ by
  <a href="https://www.codehat.de">CodeHat</a> and
  <a href="https://github.com/kodehat/portkey/graphs/contributors">
    contributors
  </a>
</div>

## Table of Contents

- [Features](#features)
- [Screenshots](#screenshots)
- [Installation](#installation)
- [Usage](#usage)
- [Configuration](#configuration)
- [Docker](#docker)
- [Development](#development)
- [See Also](#see-also)
- [License](#license)

## Features

- 🔗 Shows a collection of links acting as startup page or similiar.
- 🔎 Includes a search box with configurable keyword support.
- 📄 Can be configured easily by modifying only one file.
- 🗒️ Also supports adding smaller custom pages.
- 🌓 Dark and light mode available.
- 🪰 Very lightweight application with Docker images available.

## Screenshots

<p>You can also find a demo <a href="https://demo.portkey.page">here</a></p>

<details>

  <summary>Click to view screenshots</summary>

  #### Link collection

  ![LinkCollection](docs/images/screenshot_full.png "Link collection")
  
  #### Search for a link

  ![SearchForLink](docs/images/screenshot_search.png "Search for a link")

  #### Custom page
  
  ![CustomPage](docs/images/screenshot_custom_page.png "Custom page")

</details>

## Installation

Download the `portkey` file for your OS. Probably to a location that is in your `PATH`, so you can use it right away. 

## Usage

1. Create a `config.yml` or use the [example configuration](https://github.com/kodehat/portkey/blob/main/config.yml) from this repository and configure it as you want.
> You can find a detailed explanation of all configuration options [here](#configuration).
2. Start the application with `portkey --config-path=<path_to_config_yml>`. Providing the path to the configuration file is optional if it's in the working directory.
3. Open browser at defined host and port. Default is <http://localhost:3000>

## Configuration

The `config.yml` contains the following configuration options:

### Server

```yaml
# Set the host where the application should bind to.
host: localhost
# Set the port where the application should bind to.
port: 3000
# Set the context path (aka base-url) portkey is hosted under. Must not be specified unless you're using a reverse proxy and are hosting portkey under a directory. If that's the case then you can set this value to e.g. /portkey or whatever the directory is called. Note that the forward slash (/) in the beginning is required!
contextPath: ""
```

### Styling

```yaml
# Title of the application shown in the browser tab and on the front page.
title: "portkey"
# Allows to hide the title.
hideTitle: false
# Allows adding additional scripts/stylesheets etc. to the HTML header. Can be useful for analytics or smaller style modifications.
headerAddition: |-
  <script async src="https://analytics.example.com"></script>
# Footer (HTML support) that is shown on every page.
# Remember that Tailwind CSS classes used here do only work if already used somewhere else in the application because the bundler couldn't look here!
footer: |-
  <p>This is a footer!</p>
# Defines whether portkey's application icon should be shown at the top left of the front page.
showTopIcon: true
# If true all links are sorted alphabetically when shown on the front page. Otherwise they are shown in the order they are defined.
sortAlphabetically: false
```

### Portals (Links)

```yaml
# Defines a list of portals (links) that have additional attributes defining their appearance.
portals:
  # Name of the link
- title: example
  # (Optional) An emoji shown in front of the title.
  emoji: 🔗
  # Link where the portal will lead to (can be relative for custom pages or absolute otherwise)
  link: https://example.com/
  # If the link configured for this portal opens an external url or a relative one.
  external: true
  # Additional keywords used by the search feature.
  keywords:
  - url
  - example
```

### Custom pages

```yaml
# Defines a list of custom pages that are made available at the defined paths.
# Important: These are not automatically added to the list of portals and have to be added manually!
# This may be changed in the future.
pages:
  # Heading for the custom page. Shown in browser tab and as heading on the page.
- heading: Custom
  # Path where the custom page will be available.
  path: /custom
  # Content of the custom page and it supports using HTML.
  # The same CSS rules apply as for the footer!
  content: |-
    This is a <em>custom page</em></br>
    It also supports using <strong>HTML</strong>!
```

## Docker

There are also Docker images available at Docker hub that you can use. You can start a container for instance with:

```sh
# Assumes that there is a config.yml in the current directory.
# It is porbably better to use a specific version than 'latest'.
docker run --rm -it -v $(PWD)/config.yml:/opt/config.yml -p 3000:3000 codehat/portkey:latest
```

## Development

### Application Code

**portkey** is a *Go* application. You can install its dependencies with `go mod download`.

Live reloading is possible by installing [air](https://github.com/cosmtrek/air) and calling `air`.

### Frontend

The frontend dependencies (e.g. TailwindCSS, AlpineJS) can be installed with `npm install --include dev`.

They can be watched with `npm run watch` and built with `npm run build`.

### Templates

A library called [templ](https://templ.guide) is used for the templates. To generate the `.go` files from the templates, it has to be installed. See [here](https://templ.guide/quick-start/installation) how to do this. Afterwards you can generate the compiled templates with `templ generate`.

## See Also

The whole application is heavily inspired by a theme for the static site generator Hugo. You can find the theme at [victoriadrake/hugo-theme-sam](https://github.com/victoriadrake/hugo-theme-sam). I wanted something more dynamic while also trying out Go and improving in the language.

## License

[AGPL-3.0](https://www.tldrlegal.com/license/gnu-affero-general-public-license-v3-agpl-3-0)
# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.2.0] - 2024-10-09
### Details
#### Added
- Add Prometheus metrics by @kodehat

#### Changed
- Bump wangyoucao577/go-release-action from 1.51 to 1.52 by @dependabot[bot]
- Bump htmx.org from 2.0.2 to 2.0.3 by @dependabot[bot]
- Do not stop Fly machine by @kodehat

## [1.1.0] - 2024-09-29
### Details
#### Added
- Add Docker build script for local building by @kodehat
- Add possibility to host under a specific path by @kodehat
- Add possibility to hide the title by @kodehat
- Add searching with string similarity using levenshtein metric by @kodehat
- Add log level to tests by @kodehat
- Add log configuration keys to README by @kodehat
- Add tooltips for keywords of portals by @kodehat

#### Changed
- Update dependencies by @kodehat
- Update go version in GitHub actions by @kodehat
- Update dependencies by @kodehat
- Update dependencies by @kodehat
- Make icon in top left 'full' round by @kodehat
- Update workflows by @kodehat
- Create dependabot.yml by @kodehat
- Bump tailwindcss from 3.4.10 to 3.4.13 by @dependabot[bot]
- Bump esbuild from 0.23.1 to 0.24.0 by @dependabot[bot]
- Bump concurrently from 8.2.2 to 9.0.1 by @dependabot[bot]
- Bump husky from 9.1.5 to 9.1.6 by @dependabot[bot]
- Bump docker/build-push-action from 5 to 6 by @dependabot[bot]
- Append ignored files for Visual Studio Code to .gitignore by @kodehat
- Use log/slog for logging with configurable log level and JSON logs by @kodehat
- Show message if no search results were found by @kodehat
- Improve README.md development section by @kodehat

#### Fixed
- Version missing in Docker images by @kodehat
- Clean go.mod by @kodehat
- Fix SonarQube configuration by @kodehat
- Merge RUN instructions in Dockerfile by @kodehat
- Tooltip not closed on mobile by @kodehat

## [1.0.0] - 2024-02-26
### Details
#### Added
- Add Sonar and Husky by @kodehat
- Add type for build details by @kodehat
- Add Sonar to README and add opencontainer labels by @kodehat
- Add build Docker hub latest workflow by @kodehat
- Add test workflow by @kodehat
- Add template generation to test workflow by @kodehat
- Add application details HTML comments, add hash to CSS url by @kodehat
- Add possibility to add something to the header by @kodehat
- Add release and changelog workflows by @kodehat

#### Changed
- Update .gitignore by @kodehat
- Rename pkg dir to internal by @kodehat
- Re-order npm install parameters by @kodehat
- Do not build Docker for linux/386 by @kodehat
- Restructure code and change build script by @kodehat
- Prepare tests, restructure templates and http mux by @kodehat
- Allow hiding top icon by @kodehat
- Improve README.md by @kodehat
- Improve README.md by @kodehat
- Align step name by @kodehat
- Ignore some workflows when README.md or CHANGELOG.md is changed by @kodehat

#### Fixed
- Fix Dockerfile by @kodehat
- Fix TailwindCSS generation by @kodehat
- Fix go version in build, better align header by @kodehat
- Set page title correctly, add link to project page in footer by @kodehat

#### Removed
- Remove generated files by @kodehat
- Remove generated files by @kodehat

[1.2.0]: https://github.com/kodehat/portkey/compare/v1.1.0..v1.2.0
[1.1.0]: https://github.com/kodehat/portkey/compare/v1.0.0..v1.1.0


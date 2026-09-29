# Docker Test Container

## Purpose

This repository includes `test.dockerfile` — an Alpine-based container for running **linting** and **go vet** checks in a clean, reproducible environment.

This ensures consistent code quality across different developer machines and CI systems without needing Go or golangci-lint installed locally.

## Quick Start

### Build the image

```bash
docker build -f test.dockerfile -t bc-wallet-healthcheck:test .
```

### Run default checks (vet + lint)

```bash
docker run --rm bc-wallet-healthcheck:test
```

### Run a specific command

Override the default command by appending arguments:

```bash
# Just run go vet
docker run --rm bc-wallet-healthcheck:test go vet ./...

# Just run lint
docker run --rm bc-wallet-healthcheck:test make lint

# Dry run — check what would change
docker run --rm bc-wallet-healthcheck:test sh -c "gofmt -d . | head -20"
```

## What's Inside

| Layer | Contents |
|---|---|
| Base image | `golang:1.27-alpine` (matches `go.mod`) |
| System deps | `git`, `ca-certificates` |
| Linter | `golangci-lint v2.1.6` |
| Working dir | `/app` (project root) |
| Default cmd | `go vet ./... && make lint` |

## Limitations

- **No test execution focus**: This project currently has no `*_test.go` files, so `go test` won't do anything. The primary value here is lint consistency.
- **Network required**: `go mod download` fetches external dependencies (`bc-wallet-common-lib-errors`, `bc-wallet-common-lib-logger`). Builds must have internet access.
- **Not for runtime**: This is a tooling/devcontainer image only — there is no `Dockerfile.prod` or `Dockerfile.scratch`.

## Known Lint Issues

Some pre-existing style warnings are reported by golangci-lint (e.g. `tagalign`, `goheader` template mismatches). These are tracked separately and do not affect the Docker setup itself.

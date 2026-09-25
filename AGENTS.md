# bc-wallet-common-lib-healthcheck — Development Guidelines

## Project Overview

**bc-wallet-common-lib-healthcheck** is a Go library for managing Kubernetes-style HTTP healthcheck probes. It provides three independent probe types (startup, readiness, liveness), each running on its own HTTP server with configurable port, path, timeouts, and enable/disable flag.

| Aspect | Value |
|---|---|
| Module | `github.com/crypto-bundle/bc-wallet-common-lib-healthcheck` |
| Package | `pkg/healthcheck` |
| Go version | 1.23 |
| License | MIT NON-AI |
| Author | [@gudron (Alex V Kotelnikov)](https://github.com/gudron) |

### Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/crypto-bundle/bc-wallet-common-lib-errors` | Structured error formatting with codes |
| `github.com/crypto-bundle/bc-wallet-common-lib-logger` | Logging via log/slog factory |

---

## Repository Rules

- **Read-only default**: Always read existing code before making changes. Do not write blindly.
- **Preserve conventions**: Match existing style, naming, structure, and patterns in adjacent code.
- **No comments by default**: Only add comments when the why cannot be conveyed through naming or code structure.
- **Verify before committing**: Run `go build ./...` and `make lint` after any change.

---

## Tech Stack

- **Language**: Go 1.23+
- **Logging**: `log/slog` (structured logging), no zap dependency
- **Configuration**: `envconfig` tags on structs for environment variable injection
- **Linting**: `golangci-lint` with 80+ linters enabled, configured via `.golangci.yml`
- **Build tool**: `Makefile` with `lint` target

### golangci-lint Highlights

The linter config enforces:
- Full suite of fast linters (`fast: true`)
- Presets: bugs, comment, complexity, error, format, import, metalinter, module, performance, sql, style, test, unused
- Custom rules: `wrapcheck` ignores lib-errors wrapping helpers, `gci` orders imports as standard > localmodule > default > blank
- Disabled linters: `gofumpt`, `tagalign`, `execinquery` (legacy)

---

## Architecture

### Core Components

```
NewHTTPHealthChecker(logFactorySvc, errFmtSvc, cfgSvc)
├── probes [3]probeHTTPServer          // liveness, readiness, startup
│   ├── probeUnit (xN)                  // Each probe has its own HTTP server
│   │   ├── http.Server                 // net/http server with timeout config
│   │   └── httpHandler + middleware    // Handler stack with recovery
│   │       └── []probeService          // Registered health units
│   │           └── IsHealed(ctx) bool  // Individual health check
├── ListenAndServe(ctx)                 // Launches all probe servers as goroutines
├── AddLivenessProbeUnit(probe)         // Register a liveness health unit
├── AddReadinessProbeUnit(probe)        // Register a readiness health unit
└── AddStartupProbeUnit(probe)          // Register a startup health unit
```

### Data Flow

1. **Configuration**: User creates `*HealthcheckHTTPConfig` with embedded probe configs → passes to `NewHTTPHealthChecker`
2. **Initialization**: Factory creates one `probeUnit` per enabled probe, each with its own HTTP server
3. **Registration**: User calls `Add*Liveness/Readiness/StartupProbeUnit()` to register individual health units
4. **Request**: HTTP request arrives at any probe endpoint → handler iterates all registered units → AND-reduces `IsHealed()` results
5. **Response**: Returns HTTP 200 ("Ok") if all healthy, HTTP 418 ("Failed") if any unit is unhealthy

### Key Types

| Type | Role |
|---|---|
| `HealthcheckHTTPConfig` | Aggregator embedding `LivenessHTTPConfig`, `ReadinessHTTPConfig`, `StartupHTTPConfig` |
| `LivenessHTTPConfig` | Per-probe config for liveness probe (port :8200) |
| `ReadinessHTTPConfig` | Per-probe config for readiness probe (port :8201) |
| `StartupHTTPConfig` | Per-probe config for startup probe (port :8202) |
| `probeUnit` | Internal HTTP server instance for a single probe |
| `httpHandler` | Handles requests by aggregating all registered `probeService.IsHealed()` results |
| `middlewareRecovery` | Catches panics, logs diagnostics, returns 500 |

### Interfaces (internal, unexported)

| Interface | Methods | Used By |
|---|---|---|
| `configService` | `IsLivenessProbeEnable()`, `GetLivenessListenAddress()`, ..., `IsStartupProbeEnable()` | `httpHealthChecker`, `probeUnit` |
| `probeService` | `IsHealed(ctx)` | `httpHandler` |
| `probeHTTPServer` | `AddProbeUnit(unit)`, `ListenAndServe(ctx)` | `httpHealthChecker` |
| `loggerBuilderService` | `NewSlogNamedLoggerEntry(...)`, `NewStdLoggerEntry(...)` | `probeUnit`, `newHTPPHealthCheckerServer` |
| `errorFormatterService` | `ErrorOnly(err)`, `ErrWithCode(err, code)`, ... | `probeUnit`, `httpHealthChecker` |

---

## Code Style

### Naming Conventions

- **Exported types**: CamelCase with descriptive suffix (`HTTPConfig`, `ProbeServer`, `HTTPHandler`)
- **Internal types**: lowercase (`probeUnit`, `httpHandler`, `unitConfig`)
- **Methods matching `configService`**: `Get*` prefix for value accessors, `Is*` prefix for boolean checks
- **Interface names**: `<name>Service` suffix for service-layer abstractions (`configService`, `probeService`)

### Import Ordering (gci config)

Imports are organized into sections: `standard` > `localmodule` > `default` > `blank`

```go
import (
    "fmt"        // standard
    "time"       // standard

    "github.com/crypto-bundle/..."  // localmodule (if present)
)
```

### Struct Tags

Use `envconfig` tags for configuration injection. Format convention: `envconfig:"ENV_VAR_NAME" default:"value"`

```go
type MyConfig struct {
    Port uint `envconfig:"MY_CONFIG_PORT" default:"8080"`
}
```

Field naming inside config structs follows the pattern: `HealthCheck<Probe>HTTP<Path/Port/Timeout>` (e.g., `HealthCheckLivenessHTTPPort`).

---

## Configuration

### Environment Variables

Each probe has independent configuration via envconfig:

| Variable | Default | Description |
|---|---|---|
| `HEALTH_CHECK_LIVENESS_HTTP_PATH` | `/liveness` | HTTP path for liveness probe |
| `HEALTH_CHECK_LIVENESS_HTTP_PORT` | `8200` | Listen port for liveness probe |
| `HEALTH_CHECK_LIVENESS_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `HEALTH_CHECK_LIVENESS_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `HEALTH_CHECK_LIVENESS_ENABLED` | `true` | Enable/disable liveness probe |
| `HEALTH_CHECK_READINESS_HTTP_PATH` | `/rediness` | HTTP path for readiness probe |
| `HEALTH_CHECK_READINESS_HTTP_PORT` | `8201` | Listen port for readiness probe |
| `HEALTH_CHECK_READINESS_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `HEALTH_CHECK_READINESS_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `HEALTH_CHECK_READINESS_ENABLED` | `true` | Enable/disable readiness probe |
| `HEALTH_CHECK_STARTUP_HTTP_PATH` | `/startup` | HTTP path for startup probe |
| `HEALTH_CHECK_STARTUP_HTTP_PORT` | `8202` | Listen port for startup probe |
| `HEALTH_CHECK_STARTUP_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `HEALTH_CHECK_STARTUP_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `HEALTH_CHECK_STARTUP_ENABLED` | `true` | Enable/disable startup probe |

### Config Hierarchy

```
HealthcheckHTTPConfig               // Top-level aggregator
├── *LivenessHTTPConfig             // Liveness-specific settings
├── *ReadinessHTTPConfig            // Readiness-specific settings
└── *StartupHTTPConfig              // Startup-specific settings
```

The aggregator uses Go embedding to combine all three probe configs into a single struct. Individual probe configs implement `configService` methods.

---

## Testing

- **Current state**: No tests exist (0 test files).
- **Expected approach**: Table-driven tests, one table per public function.
- **Testing targets**:
  - `NewHTTPHealthChecker` — verify correct probe creation/enabling logic
  - `httpHandler.ServeHTTP` — verify correct status code aggregation across units
  - `*HTTPConfig.Get*Params()` — verify parameter extraction correctness
  - `middlewareRecovery.ServeHTTP` — verify panic recovery behavior

Test files should follow the naming convention `*_test.go` and live alongside source files in `pkg/healthcheck/`.

---

## CI / Quality Gates

### Makefile Targets

| Target | Command | Purpose |
|---|---|---|
| `lint` | `golangci-lint run --config .golangci.yml -v ./...` | Full lint suite |

### Pre-commit Checks

Before every commit, ensure:
1. `go build ./...` — clean compilation
2. `make lint` — no lint warnings
3. `go vet ./...` — no static analysis issues

---

## License

This project is licensed under the **MIT NON-AI License**. Key restrictions:

1. **No AI training**: The Software may not be used to train or improve machine learning algorithms (including NLP, data mining, AI)
2. **No dataset inclusion**: The Software may not be included in any ML training dataset
3. **Legal consequences**: Violations are subject to legal action

Every `.go` file must include the full license header block. Use `copyright.txt.tpl` as a template for new files.

---

## Contributing

- **Maintainer**: Alex V Kotelnikov ([@gudron](https://github.com/gudron))
- **Issues**: Report via GitHub issues
- **PRs**: Follow existing code style, include changelog entry in `CHANGELOG.md`, pass all lint checks

### Release Process

1. Update `CHANGELOG.md` with version, date, and change summary
2. Bump version in `go.mod` if needed (Go version alignment)
3. Commit: `fix:crypto-bundle/CBDL#<ticket>.<Summary>`
4. Tag release

---

## File Layout

```
.
├── doc.go                          # Root package overview (optional)
├── go.mod
├── go.sum
├── LICENSE                         # MIT NON-AI License text
├── Makefile                        # Build/lint targets
├── README.md                       # Project description
├── .agents/plans/                  # Development plans
├── copyright.txt.tpl               # License header template
├── .golangci.yml                   # Linter configuration
└── pkg/
    └── healthcheck/
        ├── common.go               # Internal interfaces (configService, probeService, etc.)
        ├── const.go                # Constants (probe indices, names, tags)
        ├── doc.go                  # Package documentation
        ├── http_config.go          # Config structs (Liveness, Readiness, Startup, Healthcheck)
        ├── http_middleware_recovery.go  # Panic recovery middleware
        ├── http_probe_handler.go   # HTTP handler (aggregates probe results)
        ├── http_probe_server.go    # probeUnit — per-probe HTTP server
        ├── http_service.go         # httpHealthChecker — coordinator (entry point)
        └── types.go                # Probe index and name constants
```

---

## Known Issues

| Issue | Location | Description |
|---|---|---|
| Typo in readiness path | `http_config.go` | Default readiness path is `/rediness` (should be `/readiness`) |
| Wrong receiver on `IsStartupProbeEnable` | `http_config.go` | Method defined on `*ReadinessHTTPConfig` instead of `*StartupHTTPConfig`; returns wrong field |
| Zero test coverage | Entire package | No `*_test.go` files exist; test infrastructure needs to be created |

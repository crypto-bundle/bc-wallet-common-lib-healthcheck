# bc-wallet-common-lib-healthcheck — Development Guidelines

## Project Overview

**bc-wallet-common-lib-healthcheck** is a Go library for managing Kubernetes-style HTTP healthcheck probes. It provides five independent probe types (startup, readiness, liveness, healthcheck, ping), each running on its own HTTP server with configurable port, path, timeouts, listen address, and enable/disable flag.

| Aspect | Value |
|---|---|
| Module | `github.com/crypto-bundle/bc-wallet-common-lib-healthcheck` |
| Package | `pkg/healthcheck` |
| Go version | 1.27 |
| License | MIT NON-AI |
| Author | [@gudron (Alex V Kotelnikov)](https://github.com/gudron) |

### Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/mailru/easyjson` | JSON serialization for probe status reports (`ProbeUnitStatusReport`, `ProbeUnitResult`) |

Notes: Error formatting is implemented locally — `ErrorOnly`, `ErrWithCode`, `ErrorNoWrap`, `ErrorCodeIsOneOf`, etc. The external `bc-wallet-common-lib-errors` package is **not** a dependency of this library.

---

## Repository Rules

- **Read-only default**: Always read existing code before making changes. Do not write blindly.
- **Preserve conventions**: Match existing style, naming, structure, and patterns in adjacent code.
- **No comments by default**: Only add comments when the why cannot be conveyed through naming or code structure.
- **Verify before committing**: Run `go build ./...` and `make lint` after any change.

---

## Tech Stack

- **Language**: Go 1.27+
- **Logging**: `*log.Logger` (standard lib logger); probes use slog-style structured tags (`slog.Key("key")`) for consistent output
- **Configuration**: `envconfig` tags on structs for environment variable injection
- **Linting**: `golangci-lint` with 80+ linters enabled, configured via `.golangci.yml`
- **Build tool**: `Makefile` with `lint`, `test`, and `test-docker` targets
- **JSON**: `github.com/mailru/easyjson` code generation via `//go:generate easyjson` directives

### golangci-lint v2 Config

| Section | Detail |
|---|---|
| Format | `version: "2"` (golangci-lint v2 schema) |
| Concurrency | 4 workers, 5 m timeout |
| Modules download | `readonly` |
| Fast linters | Enabled (`fast: true`) |
| Preset rules | `exhaustive-explicit-exhaustive-switch` |

**Top-level linters** (selected):
```yaml
linters:
  default: none
  enable:
    - asasalint, asciicheck, bidichk, bodyclose, canonicalheader
    - containedctx, contextcheck, copyloopvar, cyclop
    - decorder, depguard, dogsled, dupl, dupword, durationcheck
    - err113, errcheck, errchkjson, errname, errorlint
    - exhaustive, exhaustruct_v5, fatcontext, forbidigo
    - forcetypeassert, funlen, ginkgolinter
    - gocheckcompilerdirectives, gochecknoglobals, gochecksumtype
    - gocognit, goconst, gocritic, gocyclo, godot, godox
    - gomoddirectives, gomodguard_v2, goprintffuncname
    - gosec, gosmopolitan, govet, grouper, importas
    - inamedparam, ineffassign, interfacebloat, intrange
    - ireturn, lll, loggercheck, maintidx, makezero
    - mirror, misspell, mnd, musttag, nakedret
    - nestif, nilerr, nilnil, nlreturn, noctx, nolintlint
    - nonamedreturns, nosprintfhostport, paralleltest, perfsprint
    - prealloc, predeclared, promlinter, protogetter, reassign
    - revive, rowserrcheck, sloglint, spancheck, sqlclosecheck
    - staticcheck, tagalign, tagliatelle, testableexamples
    - testifylint, testpackage, thelper, tparallel, unconvert
    - unparam, unused, usestdlibvars, varnamelen
    - wastedassign, whitespace, wrapcheck, wsl_v5
    - zerologlint, goheader
```

**Formatter section** (v2 new feature):
```yaml
formatters:
  enable:
    - gci
    - gofmt
    - goimports
```

**Exclusion rules** (moved into `run` block in v2):
- Test files (`*_test.go`): skip `govet`, `cyclop`, `err113`, `funlen`, `gocognit`, `goconst`, `paralleltest`, `testpackage`, `exhaustruct_v5`, `gochecknoglobals`, `gosec`
- `doc.go`: skip `goheader`
- Leading `* ` source comments: skip `lll`

---

## Architecture

### Core Components

```
NewHTTPHealthChecker(errFmtSvc, logger, cfgSvc)
├── probes [5]probeHTTPServer          // liveness, readiness, startup, healthcheck, ping
│   ├── probeUnit (xN)                  // Each probe has its own HTTP server
│   │   ├── http.Server                 // net/http server with timeout config
│   │   └── handler stack               // middlewareRecovery → httpHandler
│   │       └── worker probeStatesService  // Read/write state store (SetStateMessage/SetStateError/GetStatus)
├── ListenAndServe(ctx)                 // Launches all probe servers as goroutines
├── AddLivenessProbeChecker(cfg, statesSvc)  // Register a liveness probe checker unit
├── AddRedinessProbeUnit(cfg, statesSvc)     // Register a readiness probe checker unit  
├── AddStartupProbeUnit(cfg, statesSvc)      // Register a startup probe checker unit
└── (healthcheck/ping registered internally at newHTTPHealthChecker construction time)  // Routes set up immediately; server started in ListenAndServe if enabled
```

### Data Flow

1. **Configuration**: User creates `*ProbesConfig` (aggregates all sub-configs) → passes to `NewHTTPHealthChecker`
2. **Initialization**: Factory instantiates one `probeUnit` per enabled probe at registration time, each with its own `http.Server`
3. **Registration**: User calls `AddLivenessProbeChecker()`, `AddRedinessProbeUnit()`, `AddStartupProbeUnit()` with a `probeCheckerConfigService` (read-only probe params) and `probeStatesService` (mutable state store written by external callers)
4. **External State Writes**: Application service-components (e.g., MonitorServiceComponent) call `SetStateMessage()` / `SetStateError()` on the injected `ProbeStatesUnit`. The healthcheck layer itself never mutates state.
5. **Request**: HTTP request arrives at any probe endpoint → `httpHandler.ServeHTTP()` calls `worker.GetStatus()` → returns JSON snapshot of all registered unit results
6. **Response**: Returns HTTP 200 with `{status:"...",states:{...}}` if overall status is Success; HTTP 418 with same structure if any required field failed

### Key Types

| Type | Role |
|---|---|
| `ProbesConfig` | Aggregator embedding all five sub-configs; implements `configService`; delegates to nil-safe defaults |
| `LivenessHTTPConfig` | Per-probe config for liveness probe (port :8200) |
| `ReadinessHTTPConfig` | Per-probe config for readiness probe (port :8201) |
| `StartupHTTPConfig` | Per-probe config for startup probe (port :8202) |
| `HealthcheckHTTPConfig` | Per-probe config for general healthcheck endpoint (port :8080) |
| `PingProbeHTTPConfig` | Per-probe config for ping/pong connectivity check (port :8202) |
| `unitConfig` | Internal struct carrying parameters extracted from sub-configs for server instantiation; also implements `probeCheckerConfigService` |
| `probeUnit` | Internal HTTP server instance for a single probe (holds `http.Server`, `httpHandler`, config) |
| `httpHandler` | Handles requests by calling `worker.GetStatus()` → writes JSON response via easyjson; returns 200 on success, 418 on failure |
| `middlewareRecovery` | Catches panics, logs diagnostics, returns 500 |
| `ProbeUnitStatusReport` | Top-level JSON response struct with aggregated `Status` and per-unit `States` map |
| `ProbeUnitResult` | Individual unit result with `Name`, `Message`, `Code`, and internal `err`/`isRequired` fields |
| `probeStatesUnit` | Internal worker type used by probeUnits; implements `probeStatesService` interface |
| `ProbeStatesUnit` | **Exported** — thread-safe mutable state store created via `NewProbeStateUnit(...)`, written by external callers via `SetStateMessage()` / `SetStateError()`, read-only via `GetStatus()` |

### Interfaces (internal, unexported)

| Interface | Methods | Used By |
|---|---|---|
| `configService` | `IsLivenessProbeEnable()`, `GetLivenessListenAddress()`, ..., `IsHealthcheckProbeEnable()`, ..., `IsPongProbeEnable()`, `GetPingListenAddress()` | `httpHealthChecker`, `probeUnit` |
| `probeCheckerConfigService` | `GetListenAddress()`, `GetHTTPReadTimeout()`, `GetHTTPWriteTimeout()`, `GetRequestURL()`, `GetProbeName()` | `newHTPProbeCheckerServer`, `httpHandler` |
| `probeStatesService` | `SetStateMessage(name, msg)`, `SetStateError(name, err)`, `GetStatus()` | `httpHandler` |
| `probeHTTPServer` | `ListenAndServe(ctx)` | `httpHealthChecker` |
| `errorFormatterService` | `ErrWithCode(err, code)`, `NewErrorWithCode(text, code)`, `ErrorGetCode(err)`, `ErrorCodeIsOneOf(err, codes...)`, `ErrorOnly(err, details...)`, `Errorf(err, fmt, args...)`, `NewError(details...)`, ... | `httpHandler`, `middlewareRecovery`, `probeUnit`, `probeStatesUnit` |

---

## Code Style

### Naming Conventions

- **Exported types**: CamelCase with descriptive suffix (`HTTPConfig`, `ProbeServer`, `HTTPHandler`, `ProbeUnitStatusReport`)
- **Internal types**: lowercase (`probeUnit`, `httpHandler`, `unitConfig`, `probeStatesUnit`)
- **Methods matching `configService`**: `Get*` prefix for value accessors, `Is*` prefix for boolean checks
- **Interface names**: `<name>Service` suffix for service-layer abstractions (`configService`, `probeCheckerConfigService`, `probeStatesService`)

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

For Ping/Healthcheck configs, fields use simpler names without the probe prefix (e.g., `PingProbeHTTPPath`, `HealthcheckEnabled`).

---

## Configuration

### Environment Variables

Each probe has independent configuration via envconfig:

| Variable | Default | Description |
|---|---|---|
| `PROBE_LIVENESS_HTTP_PATH` | `/liveness` | HTTP path for liveness probe |
| `PROBE_LIVENESS_HTTP_PORT` | `8200` | Listen port for liveness probe |
| `PROBE_LIVENESS_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `PROBE_LIVENESS_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `PROBE_LIVENESS_ENABLED` | `true` | Enable/disable liveness probe |
| `PROBE_READINESS_HTTP_PATH` | `/readiness` | HTTP path for readiness probe |
| `PROBE_READINESS_HTTP_PORT` | `8201` | Listen port for readiness probe |
| `PROBE_READINESS_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `PROBE_READINESS_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `PROBE_READINESS_ENABLED` | `true` | Enable/disable readiness probe |
| `PROBE_STARTUP_HTTP_PATH` | `/startup` | HTTP path for startup probe |
| `PROBE_STARTUP_HTTP_PORT` | `8202` | Listen port for startup probe |
| `PROBE_STARTUP_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `PROBE_STARTUP_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `PROBE_STARTUP_ENABLED` | `true` | Enable/disable startup probe |
| `PROBE_HEALTHCHECK_PATH` | `/health` | HTTP path for healthcheck endpoint |
| `PROBE_HEALTHCHECK_PORT` | `8080` | Listen port for healthcheck endpoint |
| `PROBE_HEALTHCHECK_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `PROBE_HEALTHCHECK_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `PROBE_HEALTHCHECK_ENABLED` | `true` | Enable/disable healthcheck endpoint |
| `PROBE_PING_HTTP_PATH` | `/ping` | HTTP path for ping endpoint |
| `PROBE_PING_HTTP_PORT` | `8202` | Listen port for ping endpoint |
| `PROBE_PING_HTTP_READ_TIMEOUT` | `5s` | HTTP read timeout |
| `PROBE_PING_HTTP_WRITE_TIMEOUT` | `10s` | HTTP write timeout |
| `PROBE_PING_ENABLED` | `true` | Enable/disable ping endpoint |

### Config Hierarchy

```
ProbesConfig                    // Top-level aggregator implementing configService
├── *LivenessHTTPConfig         // Liveness-specific settings
├── *ReadinessHTTPConfig        // Readiness-specific settings
├── *StartupHTTPConfig          // Startup-specific settings
├── *HealthcheckHTTPConfig      // Healthcheck-specific settings
└── *PingProbeHTTPConfig        // Ping-specific settings
```

The aggregator embeds all five sub-configs. Its `configService` methods delegate to the respective sub-config; if a sub-config is nil, safe defaults are returned (false, empty string, zero values). Use `NewProbesConfig()` to initialize all sub-configs with their zero/default values before parsing environment variables.

---

## Testing

- **Coverage**: Unit tests exist for `probeStatesUnit` (thread-safe state management with concurrent readers/writers, error/message lifecycle, status recalculation logic).
- **Expected approach**: Table-driven tests, one table per public function.
- **Testing targets**:
  - `NewHTTPHealthChecker` — verify correct probe creation/enabling logic
  - `httpHandler.ServeHTTP` — verify correct status code / JSON aggregation
  - `probeStatesUnit` — state mutation, concurrent access, status recalculation
  - `*HTTPConfig.Get*Params()` — verify parameter extraction correctness
  - `middlewareRecovery.ServeHTTP` — verify panic recovery behavior

Test files follow the naming convention `*_test.go` and live alongside source files in `pkg/healthcheck/`.

---

## CI / Quality Gates

### Makefile Targets

| Target | Command | Purpose |
|---|---|---|
| `lint` | `golangci-lint run --config .golangci.yml -v ./...` | Full lint suite (v2 format) |
| `test` | `go test -v -race -cover ./...` | Run unit tests with race detector |
| `test-docker` | `docker build -f test.dockerfile ... && docker run ... go test` | Run tests inside Alpine container |

### Pre-commit Checks

Before every commit, ensure:
1. `go build ./...` — clean compilation
2. `make lint` — no lint warnings
3. `go vet ./...` — no static analysis issues
4. `go generate ./...` — regenerate easyjson code if `probe_status.go` or `probe_unit_result.go` changed

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
├── doc.go                          # Root package overview
├── go.mod
├── go.sum
├── LICENSE                         # MIT NON-AI License text
├── Makefile                        # Build/lint targets
├── README.md                       # Project description
├── copyright.txt.tpl               # License header template for goheader
├── .golangci.yml                   # Linter configuration (v2 format)
├── test.dockerfile                 # Docker image for tests & CI checks
└── pkg/
    └── healthcheck/
        ├── common.go               # Internal interfaces (configService, probeCheckerConfigService, probeStatesService, errorFormatterService, probeHTTPServer)
        ├── const.go                # Constants (tags, message strings)
        ├── types.go                # Probe indices (iota), name constants, status enums, String() methods
        ├── doc.go                  # Package documentation
        ├── http_config.go          # unitConfig — parameter extractor; implements probeCheckerConfigService
        ├── http_config_probes.go   # ProbesConfig — aggregator of all five sub-configs
        ├── http_config_healthcheck.go  # HealthcheckHTTPConfig — healthcheck endpoint settings
        ├── http_config_kubernetes_liveness.go  # LivenessHTTPConfig — liveness probe settings
        ├── http_config_kubernetes_readiness.go # ReadinessHTTPConfig — readiness probe settings
        ├── http_config_ping.go           # PingProbeHTTPConfig — ping/pong connectivity check settings
        ├── http_middleware_recovery.go   # Panic recovery middleware
        ├── http_probe_handler.go         # httpHandler — request handler (JSON response via easyjson)
        ├── http_probe_server.go          # probeUnit — per-probe HTTP server
        ├── http_service.go               # httpHealthChecker — coordinator (entry point)
        ├── probe_status.go               # ProbeUnitStatusReport — top-level JSON response, easyjson gen
        ├── probe_status_easyjson.go      # Auto-generated easyjson marshalling/unmarshalling
        ├── probe_unit_result.go          # ProbeUnitResult — individual unit result, Clone(), IsHealthy()
        ├── probe_unit_result_easyjson.go # Auto-generated easyjson marshalling/unmarshalling
        ├── probe_state_unit.go           # probeStatesUnit — mutable state store, thread-safe
        └── probe_state_unit_test.go      # Unit tests for probeStatesUnit
```

---

## Known Issues

| Issue | Location | Description |
|---|---|---|
| Typo: `Rediness` spelling | `types.go`, `http_config.go`, `http_config_probes.go` | Field/index names use `Rediness*` throughout (e.g., `RedinessProbeIndex`, `RedinessProbeHTTPSrv`, `ProbeNameRediness`). Should be `Readiness`. Not a simple rename due to interface method naming contracts. |
| Broken constant name | `types.go` | `HealthCheckrobeIndex` (missing 'p') — should be `HealthCheckProbeIndex` |
| Missing case in `ProbeIndex.String()` | `types.go` | No case for `HealthCheckrobeIndex` or `PingProbeIndex` — falls through to `ProbeNameUnsupported` |
| Unmatched `ProbeUnitStatusSuccessName` | `types.go` | Set to `"startup_checker_unit"` instead of a meaningful success label |
| Missing easyjson files | `pkg/healthcheck/` | `*_easyjson.go` files generated via `go generate` — must not be edited manually; regenerate only via `go:generate` directives |

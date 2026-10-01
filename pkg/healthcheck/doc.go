// MIT NON-AI License
//
// Copyright (c) 2022-2026 Aleksei Kotelnikov(gudron2s@gmail.com)
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of the software and associated documentation files (the "Software"),
// to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
// and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions.
//
// The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
//
// In addition, the following restrictions apply:
//
// 1. The Software and any modifications made to it may not be used for the purpose of training or improving machine learning algorithms,
// including but not limited to artificial intelligence, natural language processing, or data mining. This condition applies to any derivatives,
// modifications, updates based on the Software code. Any usage of the Software in an AI-training dataset is considered a breach of this License.
//
// 2. The Software may not be included in any dataset used for training or improving machine learning algorithms,
// including but not limited to artificial intelligence, natural language processing, or data mining.
//
// 3. Any person or organization found to be in violation of these restrictions will be subject to legal action and may be held liable
// for any damages resulting from such use.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
// DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
// OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package healthcheck provides Kubernetes-style HTTP probes for application lifecycle monitoring.
//
// It manages five independent checks — each runs on its own HTTP server with configurable port,
// path, timeouts, listen address, and enable/disable flag:
//
//   - Startup    (:8202 /startup)  — verifies application has finished initializing
//   - Readiness  (:8201 /readiness) — determines whether the service can accept traffic
//   - Liveness   (:8200 /liveness) — detects if the process is frozen or stuck
//   - Health     (:8080 /health)   — aggregates the overall health of all internal components
//   - Ping       (:8202 /ping)     — lightweight connectivity check (returns pong)
//
// Each probe can be enabled or disabled independently through environment variables.
package healthcheck // import "github.com/crypto-bundle/bc-wallet-common-lib-healthcheck/pkg/healthcheck"
//
// # Design Principle
//
// The healthcheck system is organized into three isolated layers, each owning a distinct responsibility:
//
//   - **HTTPCheckerAPI** — read-only HTTP layer. Handles incoming requests for all five probes and never
//     writes state. Every request calls [ProbeStatesUnit.GetStatus()] to read the latest snapshot and returns
//     an HTTP 200 or 418 response. Public methods are defined on the type returned by [NewHTTPHealthChecker]:
//     [AddLivenessProbeChecker], [AddRedinessProbeUnit], [AddStartupProbeUnit], [ListenAndServe].
//   - **ProbeStateService** — state storage only. [ProbeStatesUnit] stores per-unit results in thread-safe maps
//     (protected by sync.RWMutex) and provides two categories of operations:
//       - Writes ← called by BusinessLogicService: SetStateMessage, SetStateError
//       - Reads → called by HTTPCheckerAPI: GetStatus, GetState, GetStates, GetStatesByNames
//     It does not know about HTTP, serves no requests, and understands no business logic.
//   - **BusinessLogicService** — any component that monitors internal dependencies
//     (MonitorServiceComponent, heartbeat handlers) implements this role. It receives a *ProbeStatesUnit
//     injected at construction time and updates probe statuses through the public write methods above.
//     The healthcheck layer has no knowledge of which specific services exist — it only interacts through
//     the interfaces documented in HTTPCheckerAPI and ProbeStateService.
//
// ```
// ┌─────────────┐  read   ┌───────────────────┐  read   ┌─────────────────────┐  write  ┌─────────────────────┐
// │  HTTPClient │────────▶│   HTTPProbeServer │────────▶│  ProbeStateService  │◀────────│ BusinessLogicService│
// │             │         │                   │         │                     │         │                     │
// │ ◄───────────┘         └───────────────────┘         │ GetStatus / Set     │         │ SetStateMessage /   │
// │ HTTP Response     (handles /liveness,               └─────────────────────╯         │ SetStateError       |
// │ 200 / 418   |       /readiness, ... endpoints)                                      └─────────────────────╯
// └─────────────┘
//
// ```
//
// Data flows unidirectionally left-to-right for requests (HTTPClient → HTTPProbeServer → ProbeStateService),
// and right-to-left for writes (BusinessLogicService → ProbeStateService). Every probe response leaves
// back to the HTTPClient via the same endpoint the request entered on.
//
// This separation keeps the healthcheck library decoupled from business logic.
// Probes report only the state recorded by whatever monitor owns that component.
//
// # Getting Started
//
// Create a config and wire probes into a checker:
//
//	cfg := healthcheck.NewProbesConfig()
//	_ = envconfig.Process("", cfg)
//
//	livenessStates   := healthcheck.NewProbeStateUnit("liveness")
//	readinessStates  := healthcheck.NewProbeStateUnit("readiness")
//	startupStates    := healthcheck.NewProbeStateUnit("startup")
//
//	checker := healthcheck.NewHTTPHealthChecker(errSvc, log, cfg)
//	checker.AddLivenessProbeChecker(livenessCfg, livenessStates)
//	checker.AddRedinessProbeUnit(readinessCfg, readinessStates)
//	checker.AddStartupProbeUnit(startupCfg, startupStates)
//	go checker.ListenAndServe(ctx)
//
// When a probe responds, the caller receives [ProbeUnitStatusReport]:
//
//	{
//	  "status": "ok",
//	  "states": {
//	    "unit_name": {
//	      "name":   "db_connection",
//	      "message": "connected",
//	      "code":   0
//	    }
//	  }
//	}
//
// # Configuration
//
// Per-probe settings use `envconfig` struct tags for injection from environment variables.
// Use [ProbesConfig] to manage all five configs together, or attach individual config structs
// ([LivenessHTTPConfig], [ReadinessHTTPConfig], etc.) when configuring a single probe.
//
// Default values:
//
//	Liveness  : port 8200, path /liveness
//	Readiness : port 8201, path /readiness
//	Startup   : port 8202, path /startup
//	Health    : port 8080, path /health
//	Ping      : port 8202, path /ping
//
// Each probe supports configurable read/write timeouts (defaults 5s read, 10s write).
//
// # Response Semantics
//
// All five probes return identical JSON structure; only the HTTP status differs:
//
//   - HTTP 200 OK — [ProbeUnitStatusReport.Status] is "ok" (case-insensitive), all required units healthy
//   - HTTP 418   — any required unit reports failure or unreachable
//
// A 418 status makes it easy to distinguish health failures from other server errors
// in load-balancer dashboards and monitoring tools.
//
// # State Management
//
// [ProbeStatesUnit] stores per-unit results as FIFO history lists and recalculates overall health whenever
// [ProbeUnitStatusReport.GetStatus] is called. Units marked as required cause a 418 response
// when they fail; optional units can be unhealthy without affecting the final status.
//
// Internally each named state holds a [probeStatesList] — a fixed-capacity buffer (default N=8)
// of historical [ProbeUnitResult] snapshots. On every call to [SetStateMessage] or [SetStateError],
// a new snapshot is appended and the oldest entry is evicted when the buffer overflows.
// [GetStatus] derives its result by reading the latest snapshot from each name's history list.
//
// Usage inside your service component:
//
//	unit.SetStateMessage("my_component", "healthy")
//	unit.SetStateError("db_pool", dbErr)
//
// Concurrent writes are safe — a sync.RWMutex protects internal maps. Unit tests cover
// simultaneous read/write scenarios under the race detector. History queries are also available:
//
//	allStates := unit.GetHistory("db_pool")      // all stored snapshots for a field
//	count     := unit.GetHistoryCount("db_pool") // number of entries in the list
//
// # Middleware & Recovery
//
// Every probe runs with a panic-recovery middleware. Unhandled panics produce an HTTP 500
// response with diagnostic details logged via slog (tags: recovery_error, recovery_time).
//
// # Errors
//
// Registration returns:
//
//   - [ErrProbeTypeNotEnabled] — the probe is disabled in configuration
//   - [ErrProbeAlreadySettled] — a probe for this index is already registered
//
// # Exported Types
//
// | Type                       | Role                                                  |
// |----------------------------|-------------------------------------------------------|
// | [ProbesConfig]             | Aggregator of all five probe configs, nil-safe        |
// | [LivenessHTTPConfig]       | Liveness-specific settings (port, path, timeouts)     |
// | [ReadinessHTTPConfig]      | Readiness-specific settings                           |
// | [StartupHTTPConfig]        | Startup-specific settings                             |
// | [HealthcheckHTTPConfig]    | General health endpoint settings                      |
// | [PingProbeHTTPConfig]      | Ping/pong connectivity check settings                 |
// | [ProbeUnitStatusReport]    | Top-level JSON response body                          |
// | [ProbeUnitResult]          | Individual unit result with Clone() and IsHealthy()   |
// | [probeStatesList]          | FIFO history buffer per named state (default N=8), stores recent ProbeUnitResult snapshots with eviction |
// | [NewProbeStateList]        | Factory for creating a probeStatesList with a configurable buffer size    |
// | [ProbeStatesUnit]          | Thread-safe mutable state store with internal FIFO history lists per field  |
// | [NewProbesConfig]          | Factory that initializes all sub-configs with defaults|
// | [NewHTTPHealthChecker]     | Constructs the coordinator that manages probe servers |
// | [NewProbeStateUnit]        | Creates a new states container with FIFO history per probe ||

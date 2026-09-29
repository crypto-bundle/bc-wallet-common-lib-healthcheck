// Package healthcheck provides HTTP healthcheck probes for Kubernetes-style lifecycle management.
//
// It manages five probe types, each with its own HTTP server, listen address, and configuration:
//   - Startup probe   (default :8202 /startup) — verifies application initialization
//   - Readiness probe (default :8201 /readiness) — indicates whether the service can accept traffic
//   - Liveness probe  (default :8200 /liveness) — detects if the application is stuck or deadlocked
//   - Healthcheck     (default :8080 /health) — general health aggregation endpoint
//   - Ping              (default :8202 /ping) — lightweight connectivity check (pong response)
//
// # State Management Principle
//
// The HTTP server of lib-healthcheck is **read-only**. State writes are performed by application
// service-components that manage health monitoring (e.g., MonitorServiceComponent). The
// [probeStatesService] implementation serves only as a transient state storage layer.
//
// ```
// HTTPProbeChecker (read) --> state management service-component <-- (write) MonitorServiceComponent
// ```
//
// This read-only contract ensures the healthcheck library remains decoupled from business logic:
// the HTTP probes never mutate state directly. They only query [probeStatesService.GetStatus] to
// report what external monitors have recorded. This separation makes it possible to update health
// state from any context (scheduled checks, dependency heartbeats, external orchestrators) without
// touching the HTTP layer.
//
// # Architecture Overview
//
// Each probe type has a dedicated config struct with [envconfig] tags for environment variable injection:
//
//   - [LivenessHTTPConfig] — liveness probe settings (PROBE_LIVENESS_*)
//   - [ReadinessHTTPConfig] — readiness probe settings (PROBE_READINESS_*)
//   - [StartupHTTPConfig] — startup probe settings (PROBE_STARTUP_*)
//   - [HealthcheckHTTPConfig] — general healthcheck endpoint settings (PROBE_HEALTHCHECK_*)
//   - [PingProbeHTTPConfig] — ping/pong connectivity check settings (PROBE_PING_*)
//
// Each standalone config struct implements getter methods for port, path, timeouts, listen address,
// and enabled state. They are self-contained and do not depend on each other.
//
// For full lifecycle management, use [ProbesConfig] which aggregates all five sub-configs.
// Its methods delegate to the respective sub-configs; if a sub-config is nil, safe defaults
// are returned (false, empty string, zero values). Use [NewProbesConfig] to initialize all
// sub-configs with their zero/default values before parsing environment variables.
//
// # Configuration Examples
//
// Approach 1 — Standalone config (single probe):
//
//	import "github.com/crypto-bundle/bc-wallet-common-lib-healthcheck/pkg/healthcheck"
//
//	cfg := &healthcheck.HealthcheckHTTPConfig{}
//	if err := env.Parse(cfg); err != nil {
//		log.Fatal(err)
//	}
//	checker := healthcheck.NewHTTPHealthChecker(errFmtSvc, logSvc, cfg)
//	checker.ListenAndServe(ctx)
//
// Approach 2 — ProbesConfig (all five probes):
//
//	cfg := healthcheck.NewProbesConfig() // initializes all sub-configs
//	if err := env.Parse(cfg); err != nil {
//		log.Fatal(err)
//	}
//	checker := healthcheck.NewHTTPHealthChecker(errFmtSvc, logSvc, cfg)
//	if err := checker.AddLivenessProbeChecker(livenessCheckerCfg, livenessStates); err != nil {
//		log.Printf("liveness: %v\n", err)
//	}
//	if err := checker.AddRedinessProbeUnit(redinessCheckerCfg, redinessStates); err != nil {
//		log.Printf("readiness: %v\n", err)
//	}
//	if err := checker.AddStartupProbeUnit(startupCheckerCfg, startupStates); err != nil {
//		log.Printf("startup: %v\n", err)
//	}
//	go checker.ListenAndServe(ctx)
//
// # Construction Flow
//
// To register probe checkers:
//
//	checker := healthcheck.NewHTTPHealthChecker(logger, errFmtSvc, cfg)
//	checker.AddLivenessProbeChecker(checkerCfg, statesSvc)
//	checker.AddRedinessProbeUnit(checkerCfg, statesSvc)
//	checker.AddStartupProbeUnit(checkerCfg, statesSvc)
//	checker.ListenAndServe(ctx)
//
// Each registration takes two parameters:
//   - checkerConfigSvc (probeCheckerConfigService) — defines where the probe connects (address, timeouts, URL, name)
//   - probeStatesSvc (probeStatesService) — write-accessible state store managed by external callers
//
// Registration returns [ErrProbeTypeNotEnabled] when the probe is disabled in config,
// and [ErrProbeAlreadySettled] when a probe for the same index has already been registered.
//
// # Environment Variables
//
// Supported envconfig variables per probe:
//
//	Liveness probe:
//		PROBE_LIVENESS_HTTP_PATH=/liveness
//		PROBE_LIVENESS_HTTP_PORT=8200
//		PROBE_LIVENESS_HTTP_READ_TIMEOUT=5s
//		PROBE_LIVENESS_HTTP_WRITE_TIMEOUT=10s
//		PROBE_LIVENESS_ENABLED=true
//
//	Readiness probe:
//		PROBE_READINESS_HTTP_PATH=/readiness
//		PROBE_READINESS_HTTP_PORT=8201
//		PROBE_READINESS_HTTP_READ_TIMEOUT=5s
//		PROBE_READINESS_HTTP_WRITE_TIMEOUT=10s
//		PROBE_READINESS_ENABLED=true
//
//	Startup probe:
//		PROBE_STARTUP_HTTP_PATH=/startup
//		PROBE_STARTUP_HTTP_PORT=8202
//		PROBE_STARTUP_HTTP_READ_TIMEOUT=5s
//		PROBE_STARTUP_HTTP_WRITE_TIMEOUT=10s
//		PROBE_STARTUP_ENABLED=true
//
//	Healthcheck endpoint:
//		PROBE_HEALTHCHECK_PATH=/health
//		PROBE_HEALTHCHECK_PORT=8080
//		PROBE_HEALTHCHECK_READ_TIMEOUT=5s
//		PROBE_HEALTHCHECK_WRITE_TIMEOUT=10s
//		PROBE_HEALTHCHECK_ENABLED=true
//
//	Ping/Pong endpoint:
//		PROBE_PING_HTTP_PATH=/startup
//		PROBE_PING_HTTP_PORT=8202
//		PROBE_PING_HTTP_READ_TIMEOUT=5s
//		PROBE_PING_HTTP_WRITE_TIMEOUT=10s
//		PROBE_PING_ENABLED=true
//
// # Response Codes
//
// All probes return the same response semantics regardless of probe type:
//   - HTTP 200 OK — body "Ok" when all registered units are healthy
//   - HTTP 418 I'm a teapot — body "Failed" when any unit reports unhealthy
//
// The 418 code distinguishes healthcheck failures from standard HTTP errors and makes it easy to
// detect failure in monitoring dashboards without relying on message parsing.
//
// # Middleware
//
// Each probe server runs with a recovery middleware that catches panics and returns 500 Internal Server Error
// with diagnostic information. Panic recovery is logged via slog with recovery_error and recovery_time tags.
//
// # Interfaces
//
// The package defines several internal interfaces that decouple the healthcheck framework from
// concrete implementations, allowing users to inject their own components:
//
//   - configService — exposes per-probe configuration getters (ports, paths, timeouts, enabled flags)
//   - probeCheckerConfigService — basic probe config for checker units (address, timeouts, URL, name)
//   - probeStatesService — SetStateMessage/SetStateError/GetStatus for tracking probe health state
//   - probeHTTPServer — ListenAndServe abstraction for probe servers
//   - loggerBuilderService — factory for slog and stdlog logger entries
//   - errorFormatterService — lib-errors wrapper for structured error handling
//   - baseConfigService — environment mode detection (IsProd, IsDev, etc.)
package healthcheck // import "github.com/crypto-bundle/bc-wallet-common-lib-healthcheck/pkg/healthcheck"

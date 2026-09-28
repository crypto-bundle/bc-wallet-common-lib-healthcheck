/*
 * MIT NON-AI License
 *
 * Copyright (c) 2022-2025 Aleksei Kotelnikov(gudron2s@gmail.com)
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of the software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions.
 *
 * The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.
 *
 * In addition, the following restrictions apply:
 *
 * 1. The Software and any modifications made to it may not be used for the purpose of training or improving machine learning algorithms,
 * including but not limited to artificial intelligence, natural language processing, or data mining. This condition applies to any derivatives,
 * modifications, or updates based on the Software code. Any usage of the Software in an AI-training dataset is considered a breach of this License.
 *
 * 2. The Software may not be included in any dataset used for training or improving machine learning algorithms,
 * including but not limited to artificial intelligence, natural language processing, or data mining.
 *
 * 3. Any person or organization found to be in violation of these restrictions will be subject to legal action and may be held liable
 * for any damages resulting from such use.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
 * DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE
 * OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 */

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
// HTTPProbeChecker (read) --> state managment service-component <-- (write) MonitorServiceComponent
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
//	if err := checker.AddLivenessProbeUnit(livenessUnit); err != nil {
//		log.Printf("liveness: %v\n", err)
//	}
//	if err := checker.AddReadinessProbeUnit(readinessUnit); err != nil {
//		log.Printf("readiness: %v\n", err)
//	}
//	if err := checker.AddStartupProbeUnit(startupUnit); err != nil {
//		log.Printf("startup: %v\n", err)
//	}
//	go checker.ListenAndServe(ctx)
//
// # Construction Flow
//
// To register health units:
//
//	checker := healthcheck.NewHTTPHealthChecker(logger, errFmtSvc, cfg)
//	checker.AddLivenessProbeUnit(unit)
//	checker.AddReadinessProbeUnit(unit)
//	checker.AddStartupProbeUnit(unit)
//	checker.ListenAndServe(ctx)
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

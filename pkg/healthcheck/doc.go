/*
 * MIT NON-AI License
 *
 * Copyright (c) 2022-2026 Aleksei Kotelnikov(gudron2s@gmail.com)
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
// It manages three probe types, each with its own HTTP server, listen address, and configuration:
//   - Startup probe   (default :8202) — verifies application initialization
//   - Readiness probe (default :8201) — indicates whether the service can accept traffic
//   - Liveness probe  (default :8200) — detects if the application is stuck or deadlocked
//
// # Configuration Types
//
// Each probe type has a dedicated config struct with envconfig tags for environment variable injection:
//
//   - LivenessHTTPConfig — liveness probe settings (HEALTH_CHECK_LIVENESS_*)
//   - ReadinessHTTPConfig — readiness probe settings (HEALTH_CHECK_READINESS_*)
//   - StartupHTTPConfig — startup probe settings (HEALTH_CHECK_STARTUP_*)
//   - HealthcheckHTTPConfig — aggregator that embeds all three probe configs
//
// Each config struct implements the configService interface by providing getter methods
// for port, path, timeouts, listen address, and enabled state.
//
// # Interfaces
//
// The package defines several internal interfaces that decouple the healthcheck framework from
// concrete implementations, allowing users to inject their own components:
//
//   - configService — exposes per-probe configuration getters (ports, paths, timeouts, enabled flags)
//   - probeService — IsHealed(ctx) returns the health status for a single health unit
//   - probeHTTPServer — internal HTTP server abstraction (AddProbeUnit, ListenAndServe)
//   - loggerBuilderService — factory for slog and stdlog logger entries
//   - errorFormatterService — lib-errors wrapper for structured error handling
//
// # Construction
//
// Create a configuration and pass it to NewHTTPHealthChecker along with logger and error formatter services:
//
//	cfg := &HealthcheckHTTPConfig{
//		LivenessHTTPConfig:  &LivenessHTTPConfig{...},
//		ReadinessHTTPConfig: &ReadinessHTTPConfig{...},
//		StartupHTTPConfig:   &StartupHTTPConfig{...},
//	}
//	checker := healthcheck.NewHTTPHealthChecker(logFactorySvc, errFmtSvc, cfg)
//	checker.ListenAndServe(ctx)
//
// To register health units:
//
//	checker.AddLivenessProbeUnit(myLivenessUnit)
//	checker.AddReadinessProbeUnit(myReadinessUnit)
//	checker.AddStartupProbeUnit(myStartupUnit)
//
// # Environment Variables
//
// Supported envconfig variables per probe:
//
//	HEALTH_CHECK_LIVENESS_HTTP_PATH=/liveness
//	HEALTH_CHECK_LIVENESS_HTTP_PORT=8200
//	HEALTH_CHECK_LIVENESS_HTTP_READ_TIMEOUT=5s
//	HEALTH_CHECK_LIVENESS_HTTP_WRITE_TIMEOUT=10s
//	HEALTH_CHECK_LIVENESS_ENABLED=true
//
//	HEALTH_CHECK_READINESS_HTTP_PATH=/readiness
//	HEALTH_CHECK_READINESS_HTTP_PORT=8201
//	HEALTH_CHECK_READINESS_HTTP_READ_TIMEOUT=5s
//	HEALTH_CHECK_READINESS_HTTP_WRITE_TIMEOUT=10s
//	HEALTH_CHECK_READINESS_ENABLED=true
//
//	HEALTH_CHECK_STARTUP_HTTP_PATH=/startup
//	HEALTH_CHECK_STARTUP_HTTP_PORT=8202
//	HEALTH_CHECK_STARTUP_HTTP_READ_TIMEOUT=5s
//	HEALTH_CHECK_STARTUP_HTTP_WRITE_TIMEOUT=10s
//	HEALTH_CHECK_STARTUP_ENABLED=true
//
// # Response Codes
//
// All probes return the same response semantics regardless of probe type:
//   - HTTP 200 OK — "Ok" when all registered units are healthy
//   - HTTP 418 I'm a teapot — "Failed" when any unit reports unhealthy
//
// The 418 code distinguishes healthcheck failures from standard HTTP errors and makes it easy to
// detect failure in monitoring dashboards without relying on message parsing.
//
// # Middleware
//
// Each probe server runs with a recovery middleware that catches panics and returns 500 Internal Server Error
// with diagnostic information. Panic recovery is logged via slog with recovery_error and recovery_time tags.
package healthcheck // import "github.com/crypto-bundle/bc-wallet-common-lib-healthcheck/pkg/healthcheck"

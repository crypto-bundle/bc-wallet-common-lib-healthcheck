// Package healthcheck provides HTTP healthcheck probes for Kubernetes-style lifecycle management.
//
// It manages three probe types, each with its own HTTP server, listen address, and configuration:
//   - Startup probe   (default :8202) — verifies application initialization
//   - Readiness probe (default :8201) — indicates whether the service can accept traffic
//   - Liveness probe  (default :8200) — detects if the application is stuck or deadlocked
//
// Each probe is independently configurable via envconfig tags (port, path, timeouts, enabled/disabled).
//
// See pkg/healthcheck for detailed API documentation.
package healthcheck // import "github.com/crypto-bundle/bc-wallet-common-lib-healthcheck/pkg/healthcheck"

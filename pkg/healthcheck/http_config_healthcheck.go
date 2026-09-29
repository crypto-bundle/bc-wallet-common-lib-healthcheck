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
// modifications, or updates based on the Software code. Any usage of the Software in an AI-training dataset is considered a breach of this License.
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

package healthcheck

import (
	"fmt"
	"time"
)

// HealthcheckHTTPConfig provides a simple single-endpoint healthcheck configuration.
// It uses its own ENV namespace (`HEALTHCHECK_*`) distinct from K8s probe namespaces.
//
// Use this when you need only one healthcheck endpoint without K8s-style
// startup/readiness/liveness separation. For full K8s lifecycle probes,
// use `K8sProbesConfig` instead.
//
// Example ENV variables:
//
//	PROBE_HEALTHCHECK_PATH=/health
//	PROBE_HEALTHCHECK_PORT=8080
//	PROBE_HEALTHCHECK_READ_TIMEOUT=5s
//	PROBE_HEALTHCHECK_WRITE_TIMEOUT=10s
//	PROBE_HEALTHCHECK_ENABLED=true
type HealthcheckHTTPConfig struct {
	HealthcheckPath         string        `default:"/health" envconfig:"PROBE_HEALTHCHECK_PATH"`
	HealthcheckPort         uint          `default:"8203"    envconfig:"PROBE_HEALTHCHECK_PORT"`
	HealthcheckReadTimeout  time.Duration `default:"5s"      envconfig:"PROBE_HEALTHCHECK_READ_TIMEOUT"`
	HealthcheckWriteTimeout time.Duration `default:"10s"     envconfig:"PROBE_HEALTHCHECK_WRITE_TIMEOUT"`
	HealthcheckEnabled      bool          `default:"true"    envconfig:"PROBE_HEALTHCHECK_ENABLED"`
}

// IsHealthcheckEnabled returns true when this healthcheck endpoint is active.
func (c *HealthcheckHTTPConfig) IsHealthcheckEnabled() bool {
	return c.HealthcheckEnabled
}

// GetListenAddress returns the listen address derived from Port.
func (c *HealthcheckHTTPConfig) GetListenAddress() string {
	return fmt.Sprintf(":%d", c.HealthcheckPort)
}

// GetRequestPath returns the HTTP request path for this endpoint.
func (c *HealthcheckHTTPConfig) GetRequestPath() string {
	return c.HealthcheckPath
}

// GetReadTimeout returns the HTTP read timeout.
func (c *HealthcheckHTTPConfig) GetReadTimeout() time.Duration {
	return c.HealthcheckReadTimeout
}

// GetWriteTimeout returns the HTTP write timeout.
func (c *HealthcheckHTTPConfig) GetWriteTimeout() time.Duration {
	return c.HealthcheckWriteTimeout
}

// GetPort returns the listen port.
func (c *HealthcheckHTTPConfig) GetPort() uint {
	return c.HealthcheckPort
}

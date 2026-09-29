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

var _ configService = (*ProbesConfig)(nil)

// ProbesConfig aggregates all three Kubernetes probe configurations.
// Use this when you need full lifecycle management: startup, readiness, and liveness.
//
// Migration: replace `&HealthcheckHTTPConfig{}` with `&ProbesConfig{}`.
type ProbesConfig struct {
	LivenessProbeHTTPSrv *LivenessHTTPConfig
	RedinessProbeHTTPSrv *ReadinessHTTPConfig
	StartupProbeHTTPSrv  *StartupHTTPConfig
	HealthcheckHTTPSrv   *HealthcheckHTTPConfig
	PingHTTPSrv          *PingProbeHTTPConfig
}

// ─── configService implementation ──────────────────────────────────────────

func (c *ProbesConfig) IsLivenessProbeEnable() bool {
	if c.LivenessProbeHTTPSrv == nil {
		return false
	}

	return c.LivenessProbeHTTPSrv.IsLivenessProbeEnable()
}

func (c *ProbesConfig) GetLivenessListenAddress() string {
	if c.LivenessProbeHTTPSrv == nil {
		return ""
	}

	return c.LivenessProbeHTTPSrv.GetLivenessListenAddress()
}

func (c *ProbesConfig) GetLivenessProbeRequestPath() string {
	if c.LivenessProbeHTTPSrv == nil {
		return ""
	}

	return c.LivenessProbeHTTPSrv.GetLivenessProbeRequestPath()
}

func (c *ProbesConfig) GetLivenessProbeReadTimeout() time.Duration {
	if c.LivenessProbeHTTPSrv == nil {
		return 0
	}

	return c.LivenessProbeHTTPSrv.GetLivenessProbeReadTimeout()
}

func (c *ProbesConfig) GetLivenessProbeWriteTimeout() time.Duration {
	if c.LivenessProbeHTTPSrv == nil {
		return 0
	}

	return c.LivenessProbeHTTPSrv.GetLivenessProbeWriteTimeout()
}

func (c *ProbesConfig) GetLivenessProbeListenPort() uint {
	if c.LivenessProbeHTTPSrv == nil {
		return 0
	}

	return c.LivenessProbeHTTPSrv.GetLivenessProbeListenPort()
}

// --- Readiness ---

func (c *ProbesConfig) IsReadinessProbeEnable() bool {
	if c.RedinessProbeHTTPSrv == nil {
		return false
	}

	return c.RedinessProbeHTTPSrv.IsReadinessProbeEnable()
}

func (c *ProbesConfig) GetReadinessListenAddress() string {
	if c.RedinessProbeHTTPSrv == nil {
		return ""
	}

	return c.RedinessProbeHTTPSrv.GetReadinessListenAddress()
}

func (c *ProbesConfig) GetReadinessProbeRequestPath() string {
	if c.RedinessProbeHTTPSrv == nil {
		return ""
	}

	return c.RedinessProbeHTTPSrv.GetReadinessProbeRequestPath()
}

func (c *ProbesConfig) GetReadinessProbeReadTimeout() time.Duration {
	if c.RedinessProbeHTTPSrv == nil {
		return 0
	}

	return c.RedinessProbeHTTPSrv.GetReadinessProbeReadTimeout()
}

func (c *ProbesConfig) GetReadinessProbeWriteTimeout() time.Duration {
	if c.RedinessProbeHTTPSrv == nil {
		return 0
	}

	return c.RedinessProbeHTTPSrv.GetReadinessProbeWriteTimeout()
}

func (c *ProbesConfig) GetReadinessProbeListenPort() uint {
	if c.RedinessProbeHTTPSrv == nil {
		return 0
	}

	return c.RedinessProbeHTTPSrv.GetReadinessProbeListenPort()
}

// --- Startup ---

func (c *ProbesConfig) IsStartupProbeEnable() bool {
	if c.StartupProbeHTTPSrv == nil {
		return false
	}

	return c.StartupProbeHTTPSrv.IsStartupProbeEnable()
}

func (c *ProbesConfig) GetStartupListenAddress() string {
	if c.StartupProbeHTTPSrv == nil {
		return ""
	}

	return c.StartupProbeHTTPSrv.GetStartupListenAddress()
}

func (c *ProbesConfig) GetStartupProbeRequestPath() string {
	if c.StartupProbeHTTPSrv == nil {
		return ""
	}

	return c.StartupProbeHTTPSrv.GetStartupProbeRequestPath()
}

func (c *ProbesConfig) GetStartupProbeReadTimeout() time.Duration {
	if c.StartupProbeHTTPSrv == nil {
		return 0
	}

	return c.StartupProbeHTTPSrv.GetStartupProbeReadTimeout()
}

func (c *ProbesConfig) GetStartupProbeWriteTimeout() time.Duration {
	if c.StartupProbeHTTPSrv == nil {
		return 0
	}

	return c.StartupProbeHTTPSrv.GetStartupProbeWriteTimeout()
}

func (c *ProbesConfig) GetStartupProbeListenPort() uint {
	if c.StartupProbeHTTPSrv == nil {
		return 0
	}

	return c.StartupProbeHTTPSrv.GetStartupProbeListenPort()
}

// Healthcheck & Pong probes — delegate to embedded configs.

func (c *ProbesConfig) IsHealthcheckProbeEnable() bool {
	if c.HealthcheckHTTPSrv == nil {
		return false
	}

	return c.HealthcheckHTTPSrv.IsHealthcheckEnabled()
}

func (c *ProbesConfig) GetHealthcheckListenAddress() string {
	if c.HealthcheckHTTPSrv == nil {
		return ""
	}

	return c.HealthcheckHTTPSrv.GetListenAddress()
}

func (c *ProbesConfig) GetHealthcheckProbeRequestPath() string {
	if c.HealthcheckHTTPSrv == nil {
		return ""
	}

	return c.HealthcheckHTTPSrv.GetRequestPath()
}

func (c *ProbesConfig) GetHealthcheckProbeReadTimeout() time.Duration {
	if c.HealthcheckHTTPSrv == nil {
		return 0
	}

	return c.HealthcheckHTTPSrv.GetReadTimeout()
}

func (c *ProbesConfig) GetHealthcheckProbeWriteTimeout() time.Duration {
	if c.HealthcheckHTTPSrv == nil {
		return 0
	}

	return c.HealthcheckHTTPSrv.GetWriteTimeout()
}

func (c *ProbesConfig) GetHealthcheckProbeListenPort() uint {
	if c.HealthcheckHTTPSrv == nil {
		return 0
	}

	return c.HealthcheckHTTPSrv.GetPort()
}

func (c *ProbesConfig) IsPongProbeEnable() bool {
	if c.PingHTTPSrv == nil {
		return false
	}

	return c.PingHTTPSrv.PingProbeEnabled
}

func (c *ProbesConfig) GetPingListenAddress() string {
	if c.PingHTTPSrv == nil {
		return ""
	}

	return fmt.Sprintf(":%d", c.PingHTTPSrv.PingProbeHTTPPort)
}

func (c *ProbesConfig) GetPingProbeRequestPath() string {
	if c.PingHTTPSrv == nil {
		return ""
	}

	return c.PingHTTPSrv.PingProbeHTTPPath
}

func (c *ProbesConfig) GetPingProbeReadTimeout() time.Duration {
	if c.PingHTTPSrv == nil {
		return 0
	}

	return c.PingHTTPSrv.PingProbeHTTPReadTimeout
}

func (c *ProbesConfig) GetPingProbeWriteTimeout() time.Duration {
	if c.PingHTTPSrv == nil {
		return 0
	}

	return c.PingHTTPSrv.PingProbeHTTPWriteTimeout
}

func (c *ProbesConfig) GetPingProbeListenPort() uint {
	if c.PingHTTPSrv == nil {
		return 0
	}

	return c.PingHTTPSrv.PingProbeHTTPPort
}

// ─── legacy helpers ─────────────────────────────────────────────────────────

// GetStartupParams builds a unitConfig from the embedded StartupHTTPConfig values.
func (c *ProbesConfig) GetStartupParams() *unitConfig {
	return &unitConfig{
		HTTPListenPort:   c.StartupProbeHTTPSrv.HealthCheckStartupHTTPPort,
		HTTPReadTimeout:  c.StartupProbeHTTPSrv.HealthCheckStartupHTTPReadTimeout,
		HTTPWriteTimeout: c.StartupProbeHTTPSrv.HealthCheckStartupHTTPWriteTimeout,
		HTTPPath:         c.StartupProbeHTTPSrv.HealthCheckStartupHTTPPath,
		ProbeName:        ProbeNameStartup,
	}
}

// GetReadinessParams builds a unitConfig from the embedded ReadinessHTTPConfig values.
func (c *ProbesConfig) GetReadinessParams() *unitConfig {
	return &unitConfig{
		HTTPListenPort:   c.RedinessProbeHTTPSrv.HealthCheckReadinessHTTPPort,
		HTTPReadTimeout:  c.RedinessProbeHTTPSrv.HealthCheckReadinessHTTPReadTimeout,
		HTTPWriteTimeout: c.RedinessProbeHTTPSrv.HealthCheckReadinessHTTPWriteTimeout,
		HTTPPath:         c.RedinessProbeHTTPSrv.HealthCheckReadinessHTTPPath,
		ProbeName:        ProbeNameRediness,
	}
}

// GetLivenessParams builds a unitConfig from the embedded LivenessHTTPConfig values.
func (c *ProbesConfig) GetLivenessParams() *unitConfig {
	return &unitConfig{
		HTTPListenPort:   c.LivenessProbeHTTPSrv.HealthCheckLivenessHTTPPort,
		HTTPReadTimeout:  c.LivenessProbeHTTPSrv.HealthCheckLivenessHTTPReadTimeout,
		HTTPWriteTimeout: c.LivenessProbeHTTPSrv.HealthCheckLivenessHTTPWriteTimeout,
		HTTPPath:         c.LivenessProbeHTTPSrv.HealthCheckLivenessHTTPPath,
		ProbeName:        ProbeNameLiveness,
	}
}

// Prepare initializes the configuration (no-op placeholder).
func (c *ProbesConfig) Prepare() error {
	return nil
}

// PrepareWith initializes the configuration from a list of config services (no-op placeholder).
func (c *ProbesConfig) PrepareWith(cfgSrvList ...interface{}) error {
	return nil
}

/*
 *
 *
 * MIT NON-AI License
 *
 * Copyright (c) 2022-2024 Aleksei Kotelnikov(gudron2s@gmail.com)
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

package healthcheck

import (
	"fmt"
	"time"
)

type LivenessHTTPConfig struct {
	HealthCheckLivenessHTTPPath         string        `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_PATH" default:"/liveness"`
	HealthCheckLivenessHTTPPort         uint          `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_PORT" default:"8200"`
	HealthCheckLivenessHTTPReadTimeout  time.Duration `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckLivenessHTTPWriteTimeout time.Duration `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckLivenessEnabled          bool          `envconfig:"HEALTH_CHECK_LIVENESS_ENABLED" default:"true"`
}

func (c *LivenessHTTPConfig) IsLivenessProbeEnable() bool {
	return c.HealthCheckLivenessEnabled
}

func (c *LivenessHTTPConfig) GetLivenessListenAddress() string {
	return fmt.Sprintf(":%d", c.HealthCheckLivenessHTTPPort)
}

func (c *LivenessHTTPConfig) GetLivenessProbeRequestPath() string {
	return c.HealthCheckLivenessHTTPPath
}

func (c *LivenessHTTPConfig) GetLivenessProbeReadTimeout() time.Duration {
	return c.HealthCheckLivenessHTTPReadTimeout
}

func (c *LivenessHTTPConfig) GetLivenessProbeWriteTimeout() time.Duration {
	return c.HealthCheckLivenessHTTPWriteTimeout
}

func (c *LivenessHTTPConfig) GetLivenessProbeListenPort() uint {
	return c.HealthCheckLivenessHTTPPort
}

type ReadinessHTTPConfig struct {
	HealthCheckReadinessHTTPPath         string        `envconfig:"HEALTH_CHECK_READINESS_HTTP_PATH" default:"/rediness"`
	HealthCheckReadinessHTTPPort         uint          `envconfig:"HEALTH_CHECK_READINESS_HTTP_PORT" default:"8201"`
	HealthCheckReadinessHTTPReadTimeout  time.Duration `envconfig:"HEALTH_CHECK_READINESS_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckReadinessHTTPWriteTimeout time.Duration `envconfig:"HEALTH_CHECK_READINESS_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckReadinessEnabled          bool          `envconfig:"HEALTH_CHECK_READINESS_ENABLED" default:"true"`
}

func (c *ReadinessHTTPConfig) IsReadinessProbeEnable() bool {
	return c.HealthCheckReadinessEnabled
}

func (c *ReadinessHTTPConfig) GetReadinessListenAddress() string {
	return fmt.Sprintf(":%d", c.HealthCheckReadinessHTTPPort)
}

func (c *ReadinessHTTPConfig) GetReadinessProbeRequestPath() string {
	return c.HealthCheckReadinessHTTPPath
}

func (c *ReadinessHTTPConfig) GetReadinessProbeReadTimeout() time.Duration {
	return c.HealthCheckReadinessHTTPReadTimeout
}

func (c *ReadinessHTTPConfig) GetReadinessProbeWriteTimeout() time.Duration {
	return c.HealthCheckReadinessHTTPWriteTimeout
}

func (c *ReadinessHTTPConfig) GetReadinessProbeListenPort() uint {
	return c.HealthCheckReadinessHTTPPort
}

type StartupHTTPConfig struct {
	HealthCheckStartupHTTPPath         string        `envconfig:"HEALTH_CHECK_STARTUP_HTTP_PATH" default:"/startup"`
	HealthCheckStartupHTTPPort         uint          `envconfig:"HEALTH_CHECK_STARTUP_HTTP_PORT" default:"8202"`
	HealthCheckStartupHTTPReadTimeout  time.Duration `envconfig:"HEALTH_CHECK_STARTUP_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckStartupHTTPWriteTimeout time.Duration `envconfig:"HEALTH_CHECK_STARTUP_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckStartupEnabled          bool          `envconfig:"HEALTH_CHECK_STARTUP_ENABLED" default:"true"`
}

func (c *ReadinessHTTPConfig) IsStartupProbeEnable() bool {
	return c.HealthCheckReadinessEnabled
}

func (c *StartupHTTPConfig) GetStartupListenAddress() string {
	return fmt.Sprintf(":%d", c.HealthCheckStartupHTTPPort)
}

func (c *StartupHTTPConfig) GetStartupProbeRequestPath() string {
	return c.HealthCheckStartupHTTPPath
}

func (c *StartupHTTPConfig) GetStartupProbeReadTimeout() time.Duration {
	return c.HealthCheckStartupHTTPReadTimeout
}

func (c *StartupHTTPConfig) GetStartupProbeWriteTimeout() time.Duration {
	return c.HealthCheckStartupHTTPWriteTimeout
}

func (c *StartupHTTPConfig) GetStartupProbeListenPort() uint {
	return c.HealthCheckStartupHTTPPort
}

type HealthcheckHTTPConfig struct {
	*LivenessHTTPConfig
	*ReadinessHTTPConfig
	*StartupHTTPConfig
}

func (c *HealthcheckHTTPConfig) GetStartupParams() *unitConfig {
	return &unitConfig{
		HTTPListenPort:   c.HealthCheckStartupHTTPPort,
		HTTPReadTimeout:  c.HealthCheckStartupHTTPReadTimeout,
		HTTPWriteTimeout: c.HealthCheckStartupHTTPWriteTimeout,
		HTTPPath:         c.HealthCheckStartupHTTPPath,
		ProbeName:        ProbeNameStartup,
	}
}

func (c *HealthcheckHTTPConfig) GetReadinessParams() *unitConfig {
	return &unitConfig{
		HTTPListenPort:   c.HealthCheckReadinessHTTPPort,
		HTTPReadTimeout:  c.HealthCheckReadinessHTTPReadTimeout,
		HTTPWriteTimeout: c.HealthCheckReadinessHTTPWriteTimeout,
		HTTPPath:         c.HealthCheckReadinessHTTPPath,
		ProbeName:        ProbeNameRediness,
	}
}

func (c *HealthcheckHTTPConfig) GetLivenessParams() *unitConfig {
	return &unitConfig{
		HTTPListenPort:   c.HealthCheckLivenessHTTPPort,
		HTTPReadTimeout:  c.HealthCheckLivenessHTTPReadTimeout,
		HTTPWriteTimeout: c.HealthCheckLivenessHTTPWriteTimeout,
		HTTPPath:         c.HealthCheckLivenessHTTPPath,
		ProbeName:        ProbeNameLiveness,
	}
}

// Prepare variables to static configuration...
func (c *HealthcheckHTTPConfig) Prepare() error {
	return nil
}

func (c *HealthcheckHTTPConfig) PrepareWith(cfgSrvList ...interface{}) error {
	return nil
}

type unitConfig struct {
	HTTPPath         string
	ProbeName        string
	HTTPListenPort   uint
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
}

func (p *unitConfig) GetListenAddress() string {
	return fmt.Sprintf(":%d", p.HTTPListenPort)
}

func (p *unitConfig) GetHTTPReadTimeout() time.Duration {
	return p.HTTPReadTimeout
}

func (p *unitConfig) GetHTTPWriteTimeout() time.Duration {
	return p.HTTPWriteTimeout
}

func (p *unitConfig) GetRequestURL() string {
	return p.HTTPPath
}

func (p *unitConfig) GetProbeName() string {
	return p.ProbeName
}

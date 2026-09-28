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

package healthcheck

import (
	"fmt"
	"time"
)

type StartupHTTPConfig struct {
	HealthCheckStartupHTTPPath         string        `envconfig:"PROBE_STARTUP_HTTP_PATH" default:"/startup"`
	HealthCheckStartupHTTPPort         uint          `envconfig:"PROBE_STARTUP_HTTP_PORT" default:"8202"`
	HealthCheckStartupHTTPReadTimeout  time.Duration `envconfig:"PROBE_STARTUP_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckStartupHTTPWriteTimeout time.Duration `envconfig:"PROBE_STARTUP_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckStartupEnabled          bool          `envconfig:"PROBE_STARTUP_ENABLED" default:"true"`
}

func (c *StartupHTTPConfig) getListenAddress() string {
	return fmt.Sprintf(":%d", c.HealthCheckStartupHTTPPort)
}

// FIXED: receiver changed from *ReadinessHTTPConfig to *StartupHTTPConfig
// FIXED: field changed from HealthCheckReadinessEnabled to HealthCheckStartupEnabled
func (c *StartupHTTPConfig) IsStartupProbeEnable() bool {
	return c.HealthCheckStartupEnabled
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

func (c *StartupHTTPConfig) GetListenAddress() string {
	return c.getListenAddress()
}

func (c *StartupHTTPConfig) GetHTTPReadTimeout() time.Duration {
	return c.HealthCheckStartupHTTPReadTimeout
}

func (c *StartupHTTPConfig) GetHTTPWriteTimeout() time.Duration {
	return c.HealthCheckStartupHTTPWriteTimeout
}

func (c *StartupHTTPConfig) GetRequestURL() string {
	return c.HealthCheckStartupHTTPPath
}

func (c *StartupHTTPConfig) GetProbeName() string {
	return ProbeNameLiveness
}

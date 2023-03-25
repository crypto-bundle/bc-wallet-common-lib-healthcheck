package healthcheck

import (
	"fmt"
	"time"
)

type LivenessHTTPConfig struct {
	HealthCheckLivenessHTTPPort         uint16        `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_PORT" default:"8200"`
	HealthCheckLivenessHTTPReadTimeout  time.Duration `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckLivenessHTTPWriteTimeout time.Duration `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckLivenessHTTPPath         string        `envconfig:"HEALTH_CHECK_LIVENESS_HTTP_PATH" default:"/liveness"`
}

type ReadinessHTTPConfig struct {
	HealthCheckReadinessHTTPPort         uint16        `envconfig:"HEALTH_CHECK_READINESS_HTTP_PORT" default:"8201"`
	HealthCheckReadinessHTTPReadTimeout  time.Duration `envconfig:"HEALTH_CHECK_READINESS_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckReadinessHTTPWriteTimeout time.Duration `envconfig:"HEALTH_CHECK_READINESS_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckReadinessHTTPPath         string        `envconfig:"HEALTH_CHECK_READINESS_HTTP_PATH" default:"/rediness"`
}

type StartupHTTPConfig struct {
	HealthCheckStartupHTTPPort         uint16        `envconfig:"HEALTH_CHECK_STARTUP_HTTP_PORT" default:"8202"`
	HealthCheckStartupHTTPReadTimeout  time.Duration `envconfig:"HEALTH_CHECK_STARTUP_HTTP_READ_TIMEOUT" default:"5s"`
	HealthCheckStartupHTTPWriteTimeout time.Duration `envconfig:"HEALTH_CHECK_STARTUP_HTTP_WRITE_TIMEOUT" default:"10s"`
	HealthCheckStartupHTTPPath         string        `envconfig:"HEALTH_CHECK_STARTUP_HTTP_PATH" default:"/startup"`
}

type HealthcheckHTTPConfig struct {
	*LivenessHTTPConfig
	*ReadinessHTTPConfig
	*StartupHTTPConfig
}

func (c *HealthcheckHTTPConfig) GetStartupParams() *unitParams {
	return &unitParams{
		HTTPListenPort:   c.HealthCheckStartupHTTPPort,
		HTTPReadTimeout:  c.HealthCheckStartupHTTPReadTimeout,
		HTTPWriteTimeout: c.HealthCheckStartupHTTPWriteTimeout,
		HTTPPath:         c.HealthCheckStartupHTTPPath,
		ProbeName:        ProbeNameStartup,
	}
}

func (c *HealthcheckHTTPConfig) GetReadinessParams() *unitParams {
	return &unitParams{
		HTTPListenPort:   c.HealthCheckReadinessHTTPPort,
		HTTPReadTimeout:  c.HealthCheckReadinessHTTPReadTimeout,
		HTTPWriteTimeout: c.HealthCheckReadinessHTTPWriteTimeout,
		HTTPPath:         c.HealthCheckReadinessHTTPPath,
		ProbeName:        ProbeNameRediness,
	}
}

func (c *HealthcheckHTTPConfig) GetLivenessParams() *unitParams {
	return &unitParams{
		HTTPListenPort:   c.HealthCheckLivenessHTTPPort,
		HTTPReadTimeout:  c.HealthCheckLivenessHTTPReadTimeout,
		HTTPWriteTimeout: c.HealthCheckLivenessHTTPWriteTimeout,
		HTTPPath:         c.HealthCheckLivenessHTTPPath,
		ProbeName:        ProbeNameLiveness,
	}
}

// Prepare variables to static configuration
func (c *HealthcheckHTTPConfig) Prepare() error {
	return nil
}

func (c *HealthcheckHTTPConfig) PrepareWith(cfgSrvList ...interface{}) error {
	return nil
}

type unitParams struct {
	HTTPListenPort   uint16
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPPath         string
	ProbeName        string
}

func (p *unitParams) GetListenAddress() string {
	return fmt.Sprintf(":%d", p.HTTPListenPort)
}

func (p *unitParams) GetHTTPReadTimeout() time.Duration {
	return p.HTTPReadTimeout
}

func (p *unitParams) GetHTTPWriteTimeout() time.Duration {
	return p.HTTPWriteTimeout
}

func (p *unitParams) GetRequestURL() string {
	return p.HTTPPath
}

func (p *unitParams) GetProbeName() string {
	return p.ProbeName
}

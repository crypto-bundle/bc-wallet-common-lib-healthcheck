package healthcheck

import (
	"fmt"
	"time"
)

// --- Shared internal type ---

// unitConfig carries parameters extracted from a probe configuration for server instantiation.
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

package healthcheck

import (
	"context"
	"time"
)

type configService interface {
	IsDebug() bool

	IsLivenessProbeEnable() bool
	GetLivenessListenAddress() string
	GetLivenessProbeRequestPath() string
	GetLivenessProbeReadTimeout() time.Duration
	GetLivenessProbeWriteTimeout() time.Duration
	GetLivenessProbeListenPort() uint

	IsReadinessProbeEnable() bool
	GetReadinessListenAddress() string
	GetReadinessProbeRequestPath() string
	GetReadinessProbeReadTimeout() time.Duration
	GetReadinessProbeWriteTimeout() time.Duration
	GetReadinessProbeListenPort() uint

	IsStartupProbeEnable() bool
	GetStartupListenAddress() string
	GetStartupProbeRequestPath() string
	GetStartupProbeReadTimeout() time.Duration
	GetStartupProbeWriteTimeout() time.Duration
	GetStartupProbeListenPort() uint
}

type probeService interface {
	IsHealed(ctx context.Context) bool
}

type probeHttpServer interface {
	AddProbeUnit(unit probeService)
	ListenAndServe(ctx context.Context) error
}

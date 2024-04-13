package healthcheck

import (
	"context"
	"errors"
	"go.uber.org/zap"
)

var (
	ErrProbeTypeNotEnabled = errors.New("healthcheck probe not enabled")
)

type httpHealthChecker struct {
	logger *zap.Logger

	probes [3]probeHttpServer // liveness, rediness, startup
}

func (s *httpHealthChecker) ListenAndServe(ctx context.Context) error {
	cancelCtx, _ := context.WithCancel(ctx)

	for _, probe := range s.probes {
		if probe == nil {
			continue
		}

		go func(probeSrv probeHttpServer) {
			err := probeSrv.ListenAndServe(cancelCtx)
			if err != nil {
				s.logger.Info("unable to start listen and server process for probe")
			}
		}(probe)
	}

	s.logger.Info("all probes successfully listen up")

	return nil
}

func (s *httpHealthChecker) AddLivenessProbeUnit(probe probeService) error {
	if s.probes[LivenessProbeIndex] == nil {
		return ErrProbeTypeNotEnabled
	}

	s.probes[LivenessProbeIndex].AddProbeUnit(probe)

	return nil
}

func (s *httpHealthChecker) AddRedinessProbeUnit(probe probeService) error {
	if s.probes[RedinessProbeIndex] == nil {
		return ErrProbeTypeNotEnabled
	}

	s.probes[RedinessProbeIndex].AddProbeUnit(probe)

	return nil
}

func (s *httpHealthChecker) AddStartupProbeUnit(probe probeService) error {
	if s.probes[StartupProbeIndex] == nil {
		return ErrProbeTypeNotEnabled
	}

	s.probes[StartupProbeIndex].AddProbeUnit(probe)

	return nil
}

func NewHTTPHealthChecker(l *zap.Logger, cfgSvc configService) *httpHealthChecker {
	probes := [3]probeHttpServer{}
	if cfgSvc.IsStartupProbeEnable() {
		probes[StartupProbeIndex] = newHTPPHealthCheckerServer(l, &unitConfig{
			HTTPListenPort:   cfgSvc.GetStartupProbeListenPort(),
			HTTPReadTimeout:  cfgSvc.GetStartupProbeReadTimeout(),
			HTTPWriteTimeout: cfgSvc.GetStartupProbeWriteTimeout(),
			HTTPPath:         cfgSvc.GetStartupProbeRequestPath(),
			ProbeName:        ProbeNameStartup,
		})
	}

	if cfgSvc.IsReadinessProbeEnable() {
		probes[RedinessProbeIndex] = newHTPPHealthCheckerServer(l, &unitConfig{
			HTTPListenPort:   cfgSvc.GetReadinessProbeListenPort(),
			HTTPReadTimeout:  cfgSvc.GetReadinessProbeReadTimeout(),
			HTTPWriteTimeout: cfgSvc.GetReadinessProbeWriteTimeout(),
			HTTPPath:         cfgSvc.GetReadinessProbeRequestPath(),
			ProbeName:        ProbeNameRediness,
		})
	}

	if cfgSvc.IsLivenessProbeEnable() {
		probes[LivenessProbeIndex] = newHTPPHealthCheckerServer(l, &unitConfig{
			HTTPListenPort:   cfgSvc.GetLivenessProbeListenPort(),
			HTTPReadTimeout:  cfgSvc.GetLivenessProbeReadTimeout(),
			HTTPWriteTimeout: cfgSvc.GetLivenessProbeWriteTimeout(),
			HTTPPath:         cfgSvc.GetLivenessProbeRequestPath(),
			ProbeName:        ProbeNameLiveness,
		})
	}

	healthChecker := &httpHealthChecker{
		logger: l,
		probes: probes,
	}

	return healthChecker
}

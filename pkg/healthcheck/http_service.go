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
	"context"
	"errors"
	"log/slog"
)

var (
	ErrProbeTypeNotEnabled = errors.New("healthcheck probe not enabled")
)

type httpHealthChecker struct {
	l *slog.Logger
	e errorFormatterService

	logFactory loggerService

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
				s.l.Error("unable to start listen and server process for probe", err)
			}
		}(probe)
	}

	s.l.Info("all probes successfully listen up")

	return nil
}

func (s *httpHealthChecker) AddLivenessProbeUnit(probe probeService) error {
	if s.probes[LivenessProbeIndex] == nil {
		return s.e.ErrorOnly(ErrProbeTypeNotEnabled)
	}

	s.probes[LivenessProbeIndex].AddProbeUnit(probe)

	return nil
}

func (s *httpHealthChecker) AddRedinessProbeUnit(probe probeService) error {
	if s.probes[RedinessProbeIndex] == nil {
		return s.e.ErrorOnly(ErrProbeTypeNotEnabled)
	}

	s.probes[RedinessProbeIndex].AddProbeUnit(probe)

	return nil
}

func (s *httpHealthChecker) AddStartupProbeUnit(probe probeService) error {
	if s.probes[StartupProbeIndex] == nil {
		return s.e.ErrorOnly(ErrProbeTypeNotEnabled)
	}

	s.probes[StartupProbeIndex].AddProbeUnit(probe)

	return nil
}

func NewHTTPHealthChecker(logFactorySvc loggerService,
	errFmtSvc errorFormatterService,
	cfgSvc configService,
) *httpHealthChecker {
	probes := [3]probeHttpServer{}
	if cfgSvc.IsStartupProbeEnable() {
		probes[StartupProbeIndex] = newHTPPHealthCheckerServer(logFactorySvc,
			errFmtSvc, &unitConfig{
				HTTPListenPort:   cfgSvc.GetStartupProbeListenPort(),
				HTTPReadTimeout:  cfgSvc.GetStartupProbeReadTimeout(),
				HTTPWriteTimeout: cfgSvc.GetStartupProbeWriteTimeout(),
				HTTPPath:         cfgSvc.GetStartupProbeRequestPath(),
				ProbeName:        ProbeNameStartup,
			})
	}

	if cfgSvc.IsReadinessProbeEnable() {
		probes[RedinessProbeIndex] = newHTPPHealthCheckerServer(logFactorySvc,
			errFmtSvc, &unitConfig{
				HTTPListenPort:   cfgSvc.GetReadinessProbeListenPort(),
				HTTPReadTimeout:  cfgSvc.GetReadinessProbeReadTimeout(),
				HTTPWriteTimeout: cfgSvc.GetReadinessProbeWriteTimeout(),
				HTTPPath:         cfgSvc.GetReadinessProbeRequestPath(),
				ProbeName:        ProbeNameRediness,
			})
	}

	if cfgSvc.IsLivenessProbeEnable() {
		probes[LivenessProbeIndex] = newHTPPHealthCheckerServer(logFactorySvc,
			errFmtSvc, &unitConfig{
				HTTPListenPort:   cfgSvc.GetLivenessProbeListenPort(),
				HTTPReadTimeout:  cfgSvc.GetLivenessProbeReadTimeout(),
				HTTPWriteTimeout: cfgSvc.GetLivenessProbeWriteTimeout(),
				HTTPPath:         cfgSvc.GetLivenessProbeRequestPath(),
				ProbeName:        ProbeNameLiveness,
			})
	}

	healthChecker := &httpHealthChecker{
		l: logFactorySvc.NewSlogNamedLoggerEntry("healthcheck"),
		e: errFmtSvc,

		probes: probes,
	}

	return healthChecker
}

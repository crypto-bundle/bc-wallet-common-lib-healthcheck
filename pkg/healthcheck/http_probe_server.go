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
	"log/slog"
	"net/http"
)

type probeUnit struct {
	l *slog.Logger
	e errorFormatterService

	cfg *unitConfig

	httpSrv      *http.Server
	probeHandler *httpHandler

	applicationPID int
}

func (s *probeUnit) ListenAndServe(ctx context.Context) error {
	err := s.httpSrv.ListenAndServe()
	if err != nil {
		s.l.Error("unable to listen and serve http server", err)

		return s.e.ErrorOnly(err)
	}

	s.l.Info("healthcheck probe server run successfully")

	<-ctx.Done()

	err = s.httpSrv.Shutdown(ctx)
	if err != nil {
		s.l.Error("unable to shutdown http server", err)

		return s.e.ErrorOnly(err)
	}

	err = s.httpSrv.Close()
	if err != nil {
		s.l.Error("unable to close http server", err)

		return s.e.ErrorOnly(err)
	}

	return nil
}

func (s *probeUnit) AddProbeUnit(unit probeService) {
	s.probeHandler.AddProbe(unit)
}

func newHTPPHealthCheckerServer(logFactorySvc loggerService,
	errFmtSvc errorFormatterService,
	configSvc *unitConfig,
) *probeUnit {
	l := logFactorySvc.NewSlogNamedLoggerEntry("healthcheck_unit",
		slog.String(ListenAddressTag, configSvc.GetListenAddress()),
		slog.String(UnitNameTag, configSvc.GetProbeName()))

	mux := http.NewServeMux()

	httpMiddleware := newMiddleware(l)
	handler := newHttpHandler()
	handlerWithMiddleware := httpMiddleware.With(handler).
		With(newRecoveryMiddleware(l))

	mux.Handle(configSvc.GetRequestURL(), handlerWithMiddleware.GetHTTPHandler())

	server := &http.Server{
		Addr:         configSvc.GetListenAddress(),
		Handler:      mux,
		ReadTimeout:  configSvc.GetHTTPReadTimeout(),
		WriteTimeout: configSvc.GetHTTPWriteTimeout(),
		ErrorLog:     logFactorySvc.NewStdLoggerEntry(),
	}

	return &probeUnit{
		l: l,
		e: errFmtSvc,

		cfg: configSvc,

		applicationPID: -1,

		httpSrv:      server,
		probeHandler: handler,
	}
}

package healthcheck

import (
	"context"
	"go.uber.org/zap"
	"net/http"
)

type probeUnit struct {
	logger *zap.Logger
	cfg    *unitConfig

	httpSrv      *http.Server
	probeHandler *httpHandler

	applicationPID int
}

func (s *probeUnit) ListenAndServe(ctx context.Context) error {
	err := s.httpSrv.ListenAndServe()
	if err != nil {
		s.logger.Error("unable to listen and serve http server", zap.Error(err))
		return err
	}

	s.logger.Info("healthcheck probe server run successfully")

	<-ctx.Done()

	err = s.httpSrv.Shutdown(ctx)
	if err != nil {
		s.logger.Error("unable to shutdown http server", zap.Error(err))
		return err
	}

	err = s.httpSrv.Close()
	if err != nil {
		s.logger.Error("unable to close http server", zap.Error(err))
		return err
	}

	return nil
}

func (s *probeUnit) AddProbeUnit(unit probeService) {
	s.probeHandler.AddProbe(unit)
}

func newHTPPHealthCheckerServer(configSvc *unitConfig,
	logger *zap.Logger,
) *probeUnit {
	l := logger.Named("healthcheck_unit").
		With(zap.String(ListenAddressTag, configSvc.GetListenAddress())).
		With(zap.String(UnitNameTag, configSvc.GetProbeName()))

	mux := http.NewServeMux()

	httpMiddleware := newMiddleware(logger)
	handler := newHttpHandler()
	handlerWithMiddleware := httpMiddleware.With(handler).
		With(newRecoveryMiddleware(logger))

	mux.Handle(configSvc.GetRequestURL(), handlerWithMiddleware.GetHTTPHandler())

	server := &http.Server{
		Addr:         configSvc.GetListenAddress(),
		Handler:      mux,
		ReadTimeout:  configSvc.GetHTTPReadTimeout(),
		WriteTimeout: configSvc.GetHTTPWriteTimeout(),
		ErrorLog:     zap.NewStdLog(logger),
	}

	return &probeUnit{
		cfg:            configSvc,
		logger:         l,
		applicationPID: -1,

		httpSrv:      server,
		probeHandler: handler,
	}
}

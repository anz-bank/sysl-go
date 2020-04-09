package core

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"regexp"
	"time"

	"github.com/anz-bank/sysl-go/common"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/anz-bank/sysl-go/metrics"
	"github.com/anz-bank/sysl-go/status"
	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

type RestManager interface {
	EnabledHandlers() []handlerinitialiser.RestHandlerInitialiser
	PublicServerConfig() *config.CommonHTTPServerConfig
}

type middlewareCollection struct {
	admin  []func(handler http.Handler) http.Handler
	public []func(handler http.Handler) http.Handler
}

func configureAdminServerListener(libraryConfig *config.LibraryConfig,
	handlers []handlerinitialiser.HandlerInitialiser,
	logger *logrus.Logger, promRegistry *prometheus.Registry,
	mWare []func(handler http.Handler) http.Handler) (func() error, error) {

	rootAdminRouter, adminRouter := configureRouters(libraryConfig.AdminServer.BasePath, mWare)

	adminTLSConfig, err := config.MakeTLSConfig(libraryConfig.AdminServer.Common.TLS)
	if err != nil {
		return nil, err
	}

	// Define meta-service endpoints:
	statusService := status.Service{
		BuildMetadata: buildMetadata,
		Config:        libraryConfig,
		Services:      handlers,
	}

	adminRouter.Route("/-", func(r chi.Router) {
		r.Route("/status", func(r chi.Router) {
			status.WireRoutes(r, &statusService)
		})
		r.Route("/metrics", func(r chi.Router) {
			r.Get("/", metrics.Handler(promRegistry).(http.HandlerFunc))
		})
		registerProfilingHandler(logger, libraryConfig, r)
	})

	listenAdmin := prepareServerListener(logger, rootAdminRouter, adminTLSConfig, libraryConfig.AdminServer)

	return listenAdmin, nil
}

func configurePublicServerListener(ctx context.Context, hl RestManager, logger *logrus.Logger, mWare []func(handler http.Handler) http.Handler) (func() error, error) {
	rootPublicRouter, publicRouter := configureRouters(hl.PublicServerConfig().BasePath, mWare)

	publicTLSConfig, err := config.MakeTLSConfig(hl.PublicServerConfig().Common.TLS)
	if err != nil {
		return nil, err
	}

	for _, h := range hl.EnabledHandlers() {
		h.WireRoutes(ctx, publicRouter)
	}

	if len(hl.EnabledHandlers()) == 0 {
		logger.Warn("No service handlers enabled by config.")
	}

	listenPublic := prepareServerListener(logger, rootPublicRouter, publicTLSConfig, hl.PublicServerConfig())

	return listenPublic, nil
}

func registerProfilingHandler(logger *logrus.Logger, cfg *config.LibraryConfig, parentRouter chi.Router) {
	if cfg.Profiling {
		logger.Infoln("Register profiling handlers")
		parentRouter.Group(func(r chi.Router) {
			r.HandleFunc("/pprof", pprof.Index)
			r.Handle("/allocs", pprof.Handler("allocs"))
			r.Handle("/block", pprof.Handler("block"))
			r.HandleFunc("/cmdline", pprof.Cmdline)
			r.Handle("/goroutine", pprof.Handler("goroutine"))
			r.Handle("/heap", pprof.Handler("heap"))
			r.Handle("/mutex", pprof.Handler("mutex"))
			r.HandleFunc("/profile", pprof.Profile)
			r.Handle("/threadcreate", pprof.Handler("threadcreate"))
			r.HandleFunc("/symbol", pprof.Symbol)
			r.HandleFunc("/trace", pprof.Trace)
		})
	} else {
		logger.Infoln("Skip register profiling handler due to profiling disabled")
	}
}

func makeListenFunc(server *http.Server, logger *logrus.Logger, cfg *config.CommonHTTPServerConfig) func() error {
	return func() error {
		if cfg.Common.TLS != nil {
			logger.Infof("TLS configuration present. Preparing to serve HTTPS for address: %s:%d%s", cfg.Common.HostName, cfg.Common.Port, cfg.BasePath)
			return server.ListenAndServeTLS("", "")
		}
		logger.Infof("no TLS configuration present. Preparing to serve HTTP for address: %s:%d%s", cfg.Common.HostName, cfg.Common.Port, cfg.BasePath)
		return server.ListenAndServe()
	}
}

func prepareServerListener(logger *logrus.Logger, rootRouter http.Handler, tlsConfig *tls.Config, httpConfig *config.CommonHTTPServerConfig) func() error {
	re := regexp.MustCompile(`TLS handshake error from .* EOF`) // Avoid spurious TLS errors from load balancer
	writer := &TLSLogFilter{logger, re}
	serverLogger := log.New(writer, "HTTPServer ", log.LstdFlags|log.Llongfile)

	server := makeNewServer(rootRouter, tlsConfig, httpConfig, serverLogger)

	listener := makeListenFunc(server, logger, httpConfig)
	logger.Infof("configured listener for address: %s:%d%s", httpConfig.Common.HostName, httpConfig.Common.Port, httpConfig.BasePath)

	return listener
}

func (m *middlewareCollection) addToBoth(h func(handler http.Handler) http.Handler) {
	m.admin = append(m.admin, h)
	m.public = append(m.public, h)
}

func prepareMiddleware(name string, logger *logrus.Logger, promRegistry *prometheus.Registry) middlewareCollection {
	result := middlewareCollection{}
	result.addToBoth(Recoverer(logger))

	result.public = append(result.public, common.TraceabilityMiddleware(logger))
	result.addToBoth(common.CoreRequestContextMiddleware(logger))

	if promRegistry != nil {
		metricsMiddleware := metrics.NewHTTPServerMetricsMiddleware(promRegistry, name, metrics.GetChiPathPattern)
		result.addToBoth(metricsMiddleware)
	}

	return result
}

func makeNewServer(router http.Handler, tlsConfig *tls.Config, serverConfig *config.CommonHTTPServerConfig, serverLogger *log.Logger) *http.Server {
	listenAddr := fmt.Sprintf("%s:%d", serverConfig.Common.HostName, serverConfig.Common.Port)
	return &http.Server{
		Addr:              listenAddr,
		Handler:           router,
		TLSConfig:         tlsConfig,
		ReadTimeout:       serverConfig.ReadTimeout,
		WriteTimeout:      serverConfig.WriteTimeout,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       5 * time.Second,
		MaxHeaderBytes:    http.DefaultMaxHeaderBytes,
		ErrorLog:          serverLogger,
	}
}

func configureRouters(basePath string, mWare []func(handler http.Handler) http.Handler) (rootRouter, router *chi.Mux) {
	rootRouter = chi.NewRouter()
	rootRouter.Use(mWare...)
	router = rootRouter.Route(basePath, nil).(*chi.Mux)

	return rootRouter, router
}

// SelectBasePath chooses between a specified base path and a dynmaically chosen one
func SelectBasePath(fromSpec, dynamic string) string {
	switch fromSpec {
	case "": // fromSpec not specified
		switch dynamic {
		case "": // dynamic not specified
			return "/"
		default:
			return dynamic
		}
	default: // fromSpec specified
		switch dynamic {
		case "": // dynamic not specified
			return fromSpec
		default:
			return dynamic
		}
	}
}

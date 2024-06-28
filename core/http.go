package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"regexp"
	"time"

	anzlog "github.com/anz-bank/sysl-go/log"
	"github.com/go-chi/chi/v5"

	"github.com/anz-bank/pkg/health"
	"github.com/anz-bank/sysl-go/config"
	"github.com/anz-bank/sysl-go/handlerinitialiser"
	"github.com/anz-bank/sysl-go/metrics"
	"github.com/anz-bank/sysl-go/status"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	defaultGracefulStopTimeout = 3 * time.Minute
)

type Manager interface {
	EnabledHandlers() []handlerinitialiser.HandlerInitialiser
	LibraryConfig() *config.LibraryConfig
	AdminServerConfig() *config.CommonHTTPServerConfig
	PublicServerConfig() *config.UpstreamConfig

	// AddAdminHTTPMiddleware can return nil if you do not have any additional middleware for the admin endpoint
	AddAdminHTTPMiddleware() func(ctx context.Context, r chi.Router)
}

func configureAdminServerListener(ctx context.Context, hl Manager, promRegistry *prometheus.Registry, healthServer *health.HTTPServer, mWare []func(handler http.Handler) http.Handler) (StoppableServer, error) {
	// validate hl manager configuration
	if hl.AdminServerConfig() == nil {
		return nil, errors.New("missing adminserverconfig")
	}
	if hl.LibraryConfig() == nil {
		return nil, errors.New("missing libraryconfig")
	}

	rootAdminRouter, adminRouter := configureRouters(hl.AdminServerConfig().BasePath, mWare)

	adminTLSConfig, err := config.MakeTLSConfig(ctx, hl.AdminServerConfig().Common.TLS)
	if err != nil {
		return nil, err
	}

	// Define meta-service endpoints:
	statusService := status.Service{
		BuildMetadata: buildMetadata,
		Config:        hl.LibraryConfig(),
		Services:      hl.EnabledHandlers(),
	}

	adminRouter.Route("/-", func(r chi.Router) {
		if hl.AddAdminHTTPMiddleware() != nil {
			hl.AddAdminHTTPMiddleware()(ctx, adminRouter)
		}
		r.Route("/status", func(r chi.Router) {
			status.WireRoutes(r, &statusService)
		})
		if promRegistry != nil {
			r.Route("/metrics", func(r chi.Router) {
				r.Get("/", metrics.Handler(promRegistry).(http.HandlerFunc))
			})
		}
		registerProfilingHandler(ctx, hl.LibraryConfig(), r)
	})
	adminRouter.Route("/", func(r chi.Router) {
		if healthServer != nil {
			healthServer.RegisterWith(r)
		}
	})

	listenAdmin := prepareServerListener(ctx, rootAdminRouter, adminTLSConfig, *hl.AdminServerConfig(), "REST Admin Server")

	return listenAdmin, nil
}

func configurePublicServerListener(ctx context.Context, hl Manager, mWare []func(handler http.Handler) http.Handler, hooks *Hooks) (StoppableServer, error) {
	rootPublicRouter, publicRouter := configureRouters("", mWare) // note basePath will be patched during the WireRoutes call below

	publicTLSConfig, err := config.MakeTLSConfig(ctx, hl.PublicServerConfig().HTTP.Common.TLS)
	if err != nil {
		return nil, err
	}

	for _, h := range hl.EnabledHandlers() {
		h.WireRoutes(ctx, publicRouter)
	}

	if len(hl.EnabledHandlers()) == 0 {
		anzlog.Info(ctx, "No service handlers enabled by config.")
	}

	prepareServerListenerFn := prepareServerListener
	if hooks != nil && hooks.StoppableServerBuilder != nil {
		prepareServerListenerFn = hooks.StoppableServerBuilder
	}

	listenPublic := prepareServerListenerFn(ctx, rootPublicRouter, publicTLSConfig, hl.PublicServerConfig().HTTP, "REST Public Server")

	return listenPublic, nil
}

func registerProfilingHandler(ctx context.Context, cfg *config.LibraryConfig, parentRouter chi.Router) {
	if cfg.Profiling {
		anzlog.Info(ctx, "Register profiling handlers")
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
		anzlog.Info(ctx, "Skip register profiling handler due to profiling disabled")
	}
}

type httpServer struct {
	ctx                 context.Context
	cfg                 config.CommonHTTPServerConfig
	server              *http.Server
	gracefulStopTimeout time.Duration
	name                string
}

func (s httpServer) Start() error {
	var err error
	if s.cfg.Common.TLS != nil {
		anzlog.Infof(s.ctx, "TLS configuration present. Preparing to serve HTTPS for address: %s:%d%s", s.cfg.Common.HostName, s.cfg.Common.Port, s.cfg.BasePath)
		err = s.server.ListenAndServeTLS("", "")
	} else {
		anzlog.Infof(s.ctx, "no TLS configuration present. Preparing to serve HTTP for address: %s:%d%s", s.cfg.Common.HostName, s.cfg.Common.Port, s.cfg.BasePath)
		err = s.server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s httpServer) GracefulStop() error {
	// If the underlying HTTP server does not have timeouts set to sufficiently small values,
	// and there are still some laggardly requests being processed, we may wait for an
	// unreasonably long time to stop gracefully. To avoid that, set a limit on the
	// maximum amount of time we're willing to wait. If we time out, give up and just do
	// a hard stop.
	var timeout time.Duration
	if s.gracefulStopTimeout != 0 {
		timeout = s.gracefulStopTimeout
	} else {
		timeout = defaultGracefulStopTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err := s.server.Shutdown(ctx)
	if err == context.DeadlineExceeded {
		anzlog.Infof(s.ctx, "warning: GracefulStop timed out for HTTP server, hard-stopping HTTP server")
		return s.server.Close()
	}
	return err
}

func (s httpServer) Stop() error {
	return s.server.Close()
}

func (s httpServer) GetName() string {
	return s.name
}

func prepareServerListener(ctx context.Context, rootRouter http.Handler, tlsConfig *tls.Config, httpConfig config.CommonHTTPServerConfig, name string) StoppableServer {
	re := regexp.MustCompile(`TLS handshake error from .* EOF`) // Avoid spurious TLS errors from load balancer
	writer := &TLSLogFilter{anzlog.GetLogger(ctx), re}
	serverLogger := log.New(writer, "HTTPServer ", log.LstdFlags|log.Llongfile)

	server := makeNewServer(ctx, rootRouter, tlsConfig, httpConfig, serverLogger)
	anzlog.Infof(ctx, "configured listener for address: %s:%d%s", httpConfig.Common.HostName, httpConfig.Common.Port, httpConfig.BasePath)
	return httpServer{
		ctx:    ctx,
		cfg:    httpConfig,
		server: server,
		name:   name,
	}
}

func makeNewServer(ctx context.Context, router http.Handler, tlsConfig *tls.Config, serverConfig config.CommonHTTPServerConfig, serverLogger *log.Logger) *http.Server {
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
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
}

func configureRouters(basePath string, mWare []func(handler http.Handler) http.Handler) (rootRouter, router *chi.Mux) {
	rootRouter = chi.NewRouter()
	rootRouter.Use(mWare...)
	if basePath == "" || basePath == "/" {
		router = rootRouter
	} else {
		router = rootRouter.Route(basePath, func(r chi.Router) {}).(*chi.Mux)
	}

	return rootRouter, router
}

// SelectBasePath chooses between a basePath from the spec or config (config takes precedence).
func SelectBasePath(fromSpec, fromCfg string) string {
	if fromCfg != "" {
		return fromCfg
	}

	if fromSpec == "" {
		return "/"
	}

	if fromSpec[0] != '/' {
		fromSpec = "/" + fromSpec
	}

	return fromSpec
}

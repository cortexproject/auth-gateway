package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/cortexproject/auth-gateway/middleware"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	AUTH           = "auth"
	UNAUTH         = "unauth"
	DefaultNetwork = "tcp"
)

type Config struct {
	HTTPRouter     *http.ServeMux
	HTTPListenPort int
	HTTPListenAddr string
	HTTPMiddleware []middleware.Interface

	UnAuthorizedHTTPRouter     *http.ServeMux
	UnAuthorizedHTTPListenAddr string
	UnAuthorizedHTTPListenPort int
	UnAuthorizedHTTPMiddleware []middleware.Interface

	ServerGracefulShutdownTimeout time.Duration
	HTTPServerReadTimeout         time.Duration
	HTTPServerWriteTimeout        time.Duration
	HTTPServerIdleTimeout         time.Duration
}

type Server struct {
	cfg           Config
	promRegistery *prometheus.Registry
	authServer    *server
	unAuthServer  *server
}

type server struct {
	http         *http.ServeMux
	httpServer   *http.Server
	httpListener net.Listener
}

func initAuthServer(cfg *Config, middlewares []middleware.Interface) (*server, error) {
	listenAddr := fmt.Sprintf("%s:%d", cfg.HTTPListenAddr, cfg.HTTPListenPort)
	httpListener, err := net.Listen(DefaultNetwork, listenAddr)
	if err != nil {
		return nil, err
	}

	var router *http.ServeMux
	if cfg.HTTPRouter != nil {
		router = cfg.HTTPRouter
	} else {
		router = http.NewServeMux()
	}

	httpMiddleware := append(middlewares, cfg.HTTPMiddleware...)
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: middleware.Merge(httpMiddleware...).Wrap(router),

		ReadTimeout:  cfg.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServerIdleTimeout,
	}

	return &server{
		http:         router,
		httpServer:   httpServer,
		httpListener: httpListener,
	}, nil
}

func initUnAuthServer(cfg *Config, middlewares []middleware.Interface) (*server, error) {
	// use any available port
	listenAddr := fmt.Sprintf("%s:%d", cfg.UnAuthorizedHTTPListenAddr, 0)
	unauthHttpListener, err := net.Listen(DefaultNetwork, listenAddr)
	if err != nil {
		return nil, err
	}

	// TODO: replace this with a log statement
	fmt.Println("Using port for /metrics, /pprof and /ready endpoints:", unauthHttpListener.Addr().(*net.TCPAddr).Port)

	var router *http.ServeMux
	if cfg.UnAuthorizedHTTPRouter != nil {
		router = cfg.UnAuthorizedHTTPRouter
	} else {
		router = http.NewServeMux()
	}
	unauthHttpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      middleware.Merge(middlewares...).Wrap(router),
		ReadTimeout:  cfg.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServerIdleTimeout,
	}

	return &server{
		http:         router,
		httpServer:   unauthHttpServer,
		httpListener: unauthHttpListener,
	}, nil
}

func (s *Server) RegisterTo(pattern string, handler http.Handler, where string) {
	switch where {
	case AUTH:
		s.authServer.http.Handle(pattern, handler)
	case UNAUTH:
		s.unAuthServer.http.Handle(pattern, handler)
	default:
		// TODO: replace this with a logger or something else
		fmt.Println("unknown")
	}
}

func New(cfg Config) (*Server, error) {
	reg := prometheus.NewRegistry()
	requestDuration := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "cortex",
			Name:      "request_duration_seconds",
			Help:      "Time (in seconds) spent serving HTTP requests.",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25, 50, 100},
		}, []string{"method", "route", "status_code", "ws"},
	)

	prometheusMiddleware := []middleware.Interface{
		middleware.Instrument{
			Duration: requestDuration,
		},
	}
	unAuthServer, _ := initUnAuthServer(&cfg, append(prometheusMiddleware, cfg.UnAuthorizedHTTPMiddleware...))
	registerMetrics(unAuthServer, reg)

	httpMiddleware := append(prometheusMiddleware, cfg.HTTPMiddleware...)
	authServer, _ := initAuthServer(&cfg, httpMiddleware)

	return &Server{
		cfg:           cfg,
		promRegistery: reg,
		authServer:    authServer,
		unAuthServer:  unAuthServer,
	}, nil
}

func (s *Server) Run() error {
	fmt.Println("server has started")
	errChan := make(chan error, 1)

	go func() {
		err := s.authServer.run()
		if err == http.ErrServerClosed {
			err = nil
		}

		select {
		case errChan <- err:
		default:
		}
	}()

	go func() {
		err := s.unAuthServer.run()
		if err == http.ErrServerClosed {
			err = nil
		}

		select {
		case errChan <- err:
		default:
		}
	}()

	return <-errChan
}

func (s *server) run() error {
	return s.httpServer.Serve(s.httpListener)
}

func (s *Server) Shutdown() {
	s.authServer.shutdown(s.cfg.ServerGracefulShutdownTimeout)
	s.unAuthServer.shutdown(s.cfg.ServerGracefulShutdownTimeout)
}

func (s *server) shutdown(gracefulShutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	s.httpServer.Shutdown(ctx)
}

func registerMetrics(srv *server, reg *prometheus.Registry) {
	srv.http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
}

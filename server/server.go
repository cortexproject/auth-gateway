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

const DefaultNetwork = "tcp"

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
	cfg Config

	PromRegistery *prometheus.Registry

	HTTP         *http.ServeMux
	HTTPServer   *http.Server
	HTTPListener net.Listener

	UnAuthorizedHTTP         *http.ServeMux
	UnAuthorizedHTTPServer   *http.Server
	UnAuthorizedHTTPListener net.Listener
}

func New(cfg Config) (*Server, error) {
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

	httpMiddleware := append(prometheusMiddleware, cfg.HTTPMiddleware...)
	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: middleware.Merge(httpMiddleware...).Wrap(router),

		ReadTimeout:  cfg.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServerIdleTimeout,
	}

	// use any available port
	listenAddr = fmt.Sprintf("%s:%d", cfg.UnAuthorizedHTTPListenAddr, 0)
	unauthHttpListener, err := net.Listen(DefaultNetwork, listenAddr)
	if err != nil {
		return nil, err
	}

	// TODO: replace this with a log statement
	fmt.Println("Using port for /metrics, /pprof and /ready endpoints:", unauthHttpListener.Addr().(*net.TCPAddr).Port)

	unauthRouter := http.NewServeMux()
	unauthRouter.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	unauthHttpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      middleware.Merge(prometheusMiddleware...).Wrap(unauthRouter),
		ReadTimeout:  cfg.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServerIdleTimeout,
	}

	return &Server{
		cfg: cfg,

		PromRegistery: reg,

		HTTP:         router,
		HTTPServer:   httpServer,
		HTTPListener: httpListener,

		UnAuthorizedHTTP:         unauthRouter,
		UnAuthorizedHTTPServer:   unauthHttpServer,
		UnAuthorizedHTTPListener: unauthHttpListener,
	}, nil
}

func (s *Server) Run() error {
	fmt.Println("server has started")
	errChan := make(chan error, 1)

	go func() {
		err := s.HTTPServer.Serve(s.HTTPListener)
		if err == http.ErrServerClosed {
			err = nil
		}

		select {
		case errChan <- err:
		default:
		}
	}()

	go func() {
		err := s.UnAuthorizedHTTPServer.Serve(s.UnAuthorizedHTTPListener)
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

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ServerGracefulShutdownTimeout)
	defer cancel()

	s.HTTPServer.Shutdown(ctx)

	ctx, cancel = context.WithTimeout(context.Background(), s.cfg.ServerGracefulShutdownTimeout)
	defer cancel()

	s.UnAuthorizedHTTPServer.Shutdown(ctx)
}

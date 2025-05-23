package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/cortexproject/auth-gateway/middleware"
	"github.com/cortexproject/auth-gateway/utils"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

const (
	AUTH              = "auth"
	UNAUTH            = "unauth"
	DefaultNetwork    = "tcp"
	DefaultAuthPort   = 80
	DefaultUnauthPort = 8081
)

type Config struct {
	HTTPRouter             *http.ServeMux
	HTTPListenPort         int
	HTTPListenAddr         string
	HTTPMiddleware         []middleware.Interface
	HTTPServerReadTimeout  time.Duration
	HTTPServerWriteTimeout time.Duration
	HTTPServerIdleTimeout  time.Duration

	UnAuthorizedHTTPRouter             *http.ServeMux
	UnAuthorizedHTTPListenAddr         string
	UnAuthorizedHTTPListenPort         int
	UnAuthorizedHTTPMiddleware         []middleware.Interface
	UnAuthorizedHTTPServerReadTimeout  time.Duration
	UnAuthorizedHTTPServerWriteTimeout time.Duration
	UnAuthorizedHTTPServerIdleTimeout  time.Duration

	ServerGracefulShutdownTimeout time.Duration
}

type Server struct {
	cfg           Config
	promRegistery *prometheus.Registry
	authServer    *server
	unAuthServer  *server
	ready         bool
}

type server struct {
	http         *http.ServeMux
	httpServer   *http.Server
	httpListener net.Listener
}

func initAuthServer(cfg *Config, middlewares []middleware.Interface) (*server, error) {
	port, err := checkPort(cfg.HTTPListenAddr, cfg.HTTPListenPort, DefaultAuthPort, DefaultNetwork)
	if err != nil {
		return nil, err
	}
	cfg.HTTPListenPort = port
	listenAddr := fmt.Sprintf("%s:%d", cfg.HTTPListenAddr, port)
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

	// These default values are the same as Cortex's server_config
	// See: https://cortexmetrics.io/docs/configuration/configuration-file/#server_config
	readTimeout, err := time.ParseDuration(cfg.HTTPServerReadTimeout.String())
	if err != nil {
		return nil, err
	}
	if readTimeout == 0 {
		readTimeout = 30 * time.Second
	}

	writeTimeout, err := time.ParseDuration(cfg.HTTPServerWriteTimeout.String())
	if err != nil {
		return nil, err
	}
	if writeTimeout == 0 {
		writeTimeout = 30 * time.Second
	}

	idleTimeout, err := time.ParseDuration(cfg.HTTPServerIdleTimeout.String())
	if err != nil {
		return nil, err
	}
	if idleTimeout == 0 {
		idleTimeout = 2 * time.Minute
	}

	httpMiddleware := append(middlewares, cfg.HTTPMiddleware...)
	httpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      middleware.Merge(httpMiddleware...).Wrap(router),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     log.New(utils.LogrusErrorWriter{}, "", 0),
	}

	return &server{
		http:         router,
		httpServer:   httpServer,
		httpListener: httpListener,
	}, nil
}

func initUnAuthServer(cfg *Config, middlewares []middleware.Interface) (*server, error) {
	port, err := checkPort(cfg.UnAuthorizedHTTPListenAddr, cfg.UnAuthorizedHTTPListenPort, DefaultUnauthPort, DefaultNetwork)
	if err != nil {
		return nil, err
	}
	cfg.UnAuthorizedHTTPListenPort = port
	listenAddr := fmt.Sprintf("%s:%d", cfg.UnAuthorizedHTTPListenAddr, cfg.UnAuthorizedHTTPListenPort)
	unauthHttpListener, err := net.Listen(DefaultNetwork, listenAddr)
	if err != nil {
		return nil, err
	}

	var router *http.ServeMux
	if cfg.UnAuthorizedHTTPRouter != nil {
		router = cfg.UnAuthorizedHTTPRouter
	} else {
		router = http.NewServeMux()
	}

	// These default values are the same as Cortex's server_config
	// See: https://cortexmetrics.io/docs/configuration/configuration-file/#server_config
	readTimeout, err := time.ParseDuration(cfg.UnAuthorizedHTTPServerReadTimeout.String())
	if err != nil {
		return nil, err
	}
	if readTimeout == 0 {
		readTimeout = 30 * time.Second
	}

	writeTimeout, err := time.ParseDuration(cfg.UnAuthorizedHTTPServerWriteTimeout.String())
	if err != nil {
		return nil, err
	}
	if writeTimeout == 0 {
		writeTimeout = 30 * time.Second
	}

	idleTimeout, err := time.ParseDuration(cfg.UnAuthorizedHTTPServerIdleTimeout.String())
	if err != nil {
		return nil, err
	}
	if idleTimeout == 0 {
		idleTimeout = 2 * time.Minute
	}

	unauthHttpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      middleware.Merge(middlewares...).Wrap(router),
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
		ErrorLog:     log.New(utils.LogrusErrorWriter{}, "", 0),
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
		logrus.Warnf("unexpected parameter: %s, valid options: %s, %s, %s not registered", where, AUTH, UNAUTH, pattern)
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
	unAuthServer, err := initUnAuthServer(&cfg, append(prometheusMiddleware, cfg.UnAuthorizedHTTPMiddleware...))
	if err != nil {
		return nil, err
	}

	httpMiddleware := append(prometheusMiddleware, cfg.HTTPMiddleware...)
	authServer, err := initAuthServer(&cfg, httpMiddleware)
	if err != nil {
		return nil, err
	}

	s := &Server{
		cfg:           cfg,
		promRegistery: reg,
		authServer:    authServer,
		unAuthServer:  unAuthServer,
		ready:         false,
	}

	registerEndpoints(unAuthServer, reg, s)

	return s, nil
}

func (s *Server) Run() error {
	logrus.Infof("the main server has started listening on %v", s.authServer.httpServer.Addr)
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

	logrus.Infof("the admin server has started listening on %v", s.unAuthServer.httpServer.Addr)
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

	s.ready = true

	return <-errChan
}

func (s *server) run() error {
	return s.httpServer.Serve(s.httpListener)
}

func (s *Server) Shutdown() {
	s.ready = false
	s.authServer.shutdown(s.cfg.ServerGracefulShutdownTimeout)
	s.unAuthServer.shutdown(s.cfg.ServerGracefulShutdownTimeout)
}

func (s *server) shutdown(gracefulShutdownTimeout time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), gracefulShutdownTimeout)
	defer cancel()

	s.httpListener.Close()
	s.httpServer.Shutdown(ctx)
}

func registerEndpoints(srv *server, reg *prometheus.Registry, serverInstance *Server) {
	srv.http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	srv.http.Handle("/ready", http.HandlerFunc(serverInstance.readyHandler))
	srv.http.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
}

func (s *Server) readyHandler(w http.ResponseWriter, r *http.Request) {
	if s.ready {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Ready!"))
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Not ready!"))
	}

}

func (s *Server) GetHTTPHandlers() (http.Handler, http.Handler) {
	return s.authServer.http, s.unAuthServer.http
}

func checkPortAvailable(addr string, port int, network string) bool {
	l, err := net.Listen(network, fmt.Sprintf("%s:%d", addr, port))
	if err != nil {
		return false
	}
	l.Close()
	return true
}

func checkPort(addr string, port int, defaultPort int, network string) (int, error) {
	p := port
	if port == 0 {
		logrus.Infof("port not specified, trying default port %d", defaultPort)
		if checkPortAvailable(addr, defaultPort, network) {
			p = defaultPort
		} else {
			return 0, fmt.Errorf("port %d is not available, please specify a port", defaultPort)
		}
	}
	return p, nil
}

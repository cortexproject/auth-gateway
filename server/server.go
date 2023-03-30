package server

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

const DefaultNetwork = "tcp"

type Config struct {
	HTTPListenNetwork string
	HTTPListenAddr    string
	HTTPListenPort    int

	ServerGracefulShutdownTimeout time.Duration
	HTTPServerReadTimeout         time.Duration
	HTTPServerWriteTimeout        time.Duration
	HTTPServerIdleTimeout         time.Duration

	Router *http.ServeMux
}

type Server struct {
	cfg Config

	HTTP         *http.ServeMux
	HTTPServer   *http.Server
	HTTPListener net.Listener
}

func New(cfg Config) (*Server, error) {
	network := cfg.HTTPListenNetwork
	if network == "" {
		network = DefaultNetwork
	}
	listenAddr := fmt.Sprintf("%s:%d", cfg.HTTPListenAddr, cfg.HTTPListenPort)
	httpListener, err := net.Listen(network, listenAddr)
	if err != nil {
		return nil, err
	}

	var router *http.ServeMux
	if cfg.Router != nil {
		router = cfg.Router
	} else {
		router = http.NewServeMux()
	}

	httpServer := &http.Server{
		Addr:    listenAddr,
		Handler: router,

		ReadTimeout:  cfg.HTTPServerReadTimeout,
		WriteTimeout: cfg.HTTPServerWriteTimeout,
		IdleTimeout:  cfg.HTTPServerIdleTimeout,
	}

	return &Server{
		cfg:          cfg,
		HTTPListener: httpListener,

		HTTP:       router,
		HTTPServer: httpServer,
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

	return <-errChan
}

func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.ServerGracefulShutdownTimeout)
	defer cancel()

	s.HTTPServer.Shutdown(ctx)
}

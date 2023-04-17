package gateway

import (
	"net/http"

	"github.com/cortexproject/auth-gateway/server"
)

type Gateway struct {
	distributorProxy   *Proxy
	queryFrontendProxy *Proxy
	srv                *server.Server
}

var defaultDistributorAPIs = []string{
	"/api/v1/push",
	"/api/prom/push",
}

var defaultQueryFrontendAPIs = []string{
	"/prometheus/api/v1/query",
	"/api/prom/api/v1/query",
	"/prometheus/api/v1/query_range",
	"/api/prom/api/v1/query_range",
	"/prometheus/api/v1/query_exemplars",
	"/api/prom/api/v1/query_exemplars",
	"/prometheus/api/v1/series",
	"/api/prom/api/v1/series",
	"/prometheus/api/v1/labels",
	"/api/prom/api/v1/labels",
	"/prometheus/api/v1/label/",
	"/api/prom/api/v1/label/",
	"/prometheus/api/v1/metadata",
	"/api/prom/api/v1/metadata",
	"/prometheus/api/v1/read",
	"/api/prom/api/v1/read",
	"/prometheus/api/v1/status/buildinfo",
	"/api/prom/api/v1/status/buildinfo",
}

func New(config *Config, srv *server.Server) (*Gateway, error) {
	distributorTimeouts := Timeouts{
		ReadTimeout:  config.Distributor.ReadTimeout,
		WriteTimeout: config.Distributor.WriteTimeout,
		IdleTimeout:  config.Distributor.IdleTimeout,
	}
	distributor, err := NewProxy(config.Distributor.URL, distributorTimeouts, DISTRIBUTOR)
	if err != nil {
		return nil, err
	}

	frontend, err := NewProxy(config.QueryFrontend.URL, Timeouts{}, "")
	if err != nil {
		return nil, err
	}

	return &Gateway{
		distributorProxy:   distributor,
		queryFrontendProxy: frontend,
		srv:                srv,
	}, nil
}

func (g *Gateway) Start(config *Config) {
	g.registerRoutes(config)
}

func (g *Gateway) registerRoutes(config *Config) {
	g.registerProxyRoutes(config.Distributor.Paths, defaultDistributorAPIs, http.HandlerFunc(g.distributorProxy.Handler))
	g.registerProxyRoutes(config.QueryFrontend.Paths, defaultQueryFrontendAPIs, http.HandlerFunc(g.queryFrontendProxy.Handler))
	g.srv.RegisterTo("/", http.HandlerFunc(g.notFoundHandler), server.UNAUTH)
}

func (g *Gateway) registerProxyRoutes(paths []string, defaultPaths []string, handler http.Handler) {
	pathsToRegister := defaultPaths
	if len(paths) > 0 {
		pathsToRegister = paths
	}

	for _, path := range pathsToRegister {
		g.srv.RegisterTo(path, handler, server.AUTH)
	}
}

func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("404 - Resource not found"))
}

package gateway

import (
	"net/http"

	"github.com/cortexproject/auth-gateway/server"
)

type Gateway struct {
	DistributorProxy   *Proxy
	queryFrontendProxy *Proxy
	Server             *server.Server
}

var DefaultDistributorAPIs = []string{
	"/api/v1/push",
	"/api/prom/push",
}

var DefaultQueryFrontendAPIs = []string{
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
	distributor, err := NewProxy(config.Distributor.URL)
	if err != nil {
		return nil, err
	}

	frontend, err := NewProxy(config.QueryFrontend.URL)
	if err != nil {
		return nil, err
	}

	return &Gateway{
		DistributorProxy:   distributor,
		queryFrontendProxy: frontend,
		Server:             srv,
	}, nil
}

func (g *Gateway) Start(config *Config) {
	g.registerRoutes(config)
}

func (g *Gateway) registerRoutes(config *Config) {
	g.registerProxyRoutes(config, config.Distributor.Paths, DefaultDistributorAPIs, http.HandlerFunc(g.DistributorProxy.Handler))
	g.registerProxyRoutes(config, config.QueryFrontend.Paths, DefaultQueryFrontendAPIs, http.HandlerFunc(g.queryFrontendProxy.Handler))
	g.Server.HTTP.Handle("/", http.HandlerFunc(g.notFoundHandler))
}

func (g *Gateway) registerProxyRoutes(config *Config, paths []string, defaultPaths []string, handler http.Handler) {
	pathsToRegister := defaultPaths
	if len(paths) > 0 {
		pathsToRegister = paths
	}

	for _, path := range pathsToRegister {
		g.Server.HTTP.Handle(path, config.Authenticate(handler))
	}
}

func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("404 - Resource not found"))
}

package gateway

import (
	"net/http"

	"github.com/cortexproject/auth-gateway/server"
)

type Gateway struct {
	distributorProxy *Proxy
	server           *server.Server
}

func New(config Config, srv *server.Server) (*Gateway, error) {
	distributor, err := NewProxy(config.Distributor.URL, "distributor")
	if err != nil {
		return nil, err
	}

	return &Gateway{
		distributorProxy: distributor,
		server:           srv,
	}, nil
}

func (g *Gateway) Start(config *Config) {
	g.registerRoutes(config)
}

func (g *Gateway) registerRoutes(config *Config) {
	g.server.HTTP.Handle("/api/v1/push", config.Authenticate(http.HandlerFunc(g.distributorProxy.Handler)))
	g.server.HTTP.Handle("/api/prom/push", config.Authenticate(http.HandlerFunc(g.distributorProxy.Handler)))
	g.server.HTTP.Handle("/", http.HandlerFunc(g.notFoundHandler))
}

func (g *Gateway) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
	w.Write([]byte("404 - Resource not found"))
}

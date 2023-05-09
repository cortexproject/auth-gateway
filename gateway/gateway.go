package gateway

import (
	"net/http"

	"github.com/cortexproject/auth-gateway/server"
	"github.com/sirupsen/logrus"
)

type Gateway struct {
	distributorProxy   *Proxy
	queryFrontendProxy *Proxy
	alertmanagerProxy  *Proxy
	rulerProxy         *Proxy
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

var defaultAlertmanagerAPIs = []string{
	"/alertmanager/",
	"/api/v1/alerts",
	"/multitenant_alertmanager/delete_tenant_config",
}

var defaultRulerAPIs = []string{
	"/prometheus/api/v1/rules",
	"/api/prom/api/v1/rules",
	"/prometheus/api/v1/alerts",
	"/api/prom/api/v1/alerts",
	"/api/v1/rules",
	"/api/v1/rules/",
	"/api/prom/rules/",
	"/ruler/delete_tenant_config",
}

func New(config *Config, srv *server.Server) (*Gateway, error) {
	gateway := &Gateway{
		srv: srv,
	}

	components := []string{DISTRIBUTOR, FRONTEND, ALERTMANAGER, RULER}
	for _, componentName := range components {
		upstreamConfig := config.getUpstreamConfig(componentName)
		proxy, err := setupProxy(upstreamConfig, componentName, componentName)
		if err != nil {
			return nil, err
		}
		switch componentName {
		case DISTRIBUTOR:
			gateway.distributorProxy = proxy
		case FRONTEND:
			gateway.queryFrontendProxy = proxy
		case ALERTMANAGER:
			gateway.alertmanagerProxy = proxy
		case RULER:
			gateway.rulerProxy = proxy
		}
	}

	return gateway, nil
}

func (g *Gateway) Start(config *Config) {
	g.registerRoutes(config)
}

func (c *Config) getUpstreamConfig(componentName string) Upstream {
	switch componentName {
	case DISTRIBUTOR:
		return c.Distributor
	case FRONTEND:
		return c.QueryFrontend
	case ALERTMANAGER:
		return c.Alertmanager
	case RULER:
		return c.Ruler
	default:
		return Upstream{}
	}
}

func setupProxy(upstreamConfig Upstream, proxyType string, description string) (*Proxy, error) {
	if upstreamConfig.URL != "" {
		proxy, err := NewProxy(upstreamConfig.URL, upstreamConfig, proxyType)
		if err != nil {
			return nil, err
		}
		return proxy, nil
	}
	logrus.Infof("%s URL configuration not provided. %s will not be set up.", description, description)
	return nil, nil
}

func (g *Gateway) registerRoutes(config *Config) {
	g.registerProxyRoutes(config.Distributor.Paths, defaultDistributorAPIs, http.HandlerFunc(g.distributorProxy.Handler))
	g.registerProxyRoutes(config.QueryFrontend.Paths, defaultQueryFrontendAPIs, http.HandlerFunc(g.queryFrontendProxy.Handler))
	g.registerProxyRoutes(config.Alertmanager.Paths, defaultAlertmanagerAPIs, http.HandlerFunc(g.alertmanagerProxy.Handler))
	g.registerProxyRoutes(config.Ruler.Paths, defaultRulerAPIs, http.HandlerFunc(g.rulerProxy.Handler))
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
	w.WriteHeader(http.StatusNotFound)
	w.Write([]byte("404 - Resource not found"))
}

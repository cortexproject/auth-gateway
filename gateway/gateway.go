package gateway

type Gateway struct {
	distributorProxy *Proxy
	// server           *server.Server
}

// This `srv *server.Server` will be taken as a parameter
func New(config Configuration) (*Gateway, error) {
	distributor, err := NewProxy(config.Targets["distributor"], "distributor")
	if err != nil {
		return nil, err
	}

	return &Gateway{
		distributorProxy: distributor,
		// server:           srv,
	}, nil
}

func (g *Gateway) Start(config *Configuration) {
	g.registerRoutes(config)
}

func (g *Gateway) registerRoutes(config *Configuration) {
	// Since the server is not implemented, below line does not compile. That is why it is commented out
	// g.server.HTTP.Handle("/api/v1/push", config.Authenticate(http.HandlerFunc(g.distributorProxy.Handler)))
}

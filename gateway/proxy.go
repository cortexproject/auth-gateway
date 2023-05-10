package gateway

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/cortexproject/auth-gateway/utils"
)

const (
	DISTRIBUTOR  = "distributor"
	FRONTEND     = "frontend"
	ALERTMANAGER = "alertmanager"
	RULER        = "ruler"
)

var defaultTimeoutValues map[string]Upstream = map[string]Upstream{
	DISTRIBUTOR: {
		HTTPClientTimeout:               time.Second * 15,
		HTTPClientDialerTimeout:         time.Second * 5,
		HTTPClientTLSHandshakeTimeout:   time.Second * 5,
		HTTPClientResponseHeaderTimeout: time.Second * 5,
	},
	FRONTEND: {
		HTTPClientTimeout:               time.Minute * 1,
		HTTPClientDialerTimeout:         time.Second * 5,
		HTTPClientTLSHandshakeTimeout:   time.Second * 5,
		HTTPClientResponseHeaderTimeout: time.Second * 5,
	},
	ALERTMANAGER: {
		HTTPClientTimeout:               time.Second * 15,
		HTTPClientDialerTimeout:         time.Second * 5,
		HTTPClientTLSHandshakeTimeout:   time.Second * 5,
		HTTPClientResponseHeaderTimeout: time.Second * 5,
	},
	RULER: {
		HTTPClientTimeout:               time.Second * 15,
		HTTPClientDialerTimeout:         time.Second * 5,
		HTTPClientTLSHandshakeTimeout:   time.Second * 5,
		HTTPClientResponseHeaderTimeout: time.Second * 5,
	},
}

type Proxy struct {
	targetURL    *url.URL
	upstream     Upstream
	reverseProxy *httputil.ReverseProxy
}

func NewProxy(targetURL string, upstream Upstream, component string) (*Proxy, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	if url.Scheme == "" {
		return nil, fmt.Errorf("invalid URL scheme when creating a proxy for the %s: %s", component, targetURL)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(url)
	transport, err := customTransport(component, upstream)
	if err != nil {
		return nil, err
	}
	reverseProxy.Transport = transport
	originalDirector := reverseProxy.Director
	reverseProxy.Director = customDirector(url, originalDirector)
	reverseProxy.ErrorLog = log.New(utils.LogrusErrorWriter{}, "", 0)

	if upstream.HTTPClientTimeout == 0 {
		upstream.HTTPClientTimeout = defaultTimeoutValues[component].HTTPClientTimeout
	}

	return &Proxy{
		targetURL:    url,
		upstream:     upstream,
		reverseProxy: reverseProxy,
	}, nil
}

// it may seem redundant right now but I plan to use this later
func customDirector(targetURL *url.URL, originalDirector func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		originalDirector(r)
	}
}

func customTransport(component string, upstream Upstream) (http.RoundTripper, error) {
	dialerTimeout := upstream.HTTPClientDialerTimeout * time.Second
	if dialerTimeout == 0 {
		dialerTimeout = defaultTimeoutValues[component].HTTPClientDialerTimeout
	}
	TLSHandshakeTimeout := upstream.HTTPClientTLSHandshakeTimeout * time.Second
	if TLSHandshakeTimeout == 0 {
		TLSHandshakeTimeout = defaultTimeoutValues[component].HTTPClientTLSHandshakeTimeout
	}
	responseHeaderTimeout := upstream.HTTPClientResponseHeaderTimeout * time.Second
	if responseHeaderTimeout == 0 {
		responseHeaderTimeout = defaultTimeoutValues[component].HTTPClientResponseHeaderTimeout
	}
	dnsRefreshInterval := upstream.DNSRefreshInterval * time.Second
	if dnsRefreshInterval == 0 {
		dnsRefreshInterval = 1 * time.Second
	}

	url, err := url.Parse(upstream.URL)
	if err != nil {
		return nil, fmt.Errorf("unexpected error when parsing the upstream url: %v", err)
	}

	resolver := DefaultDNSResolver{}
	lb, err := newRoundRobinLoadBalancer(url.Hostname(), resolver.LookupIP)
	if err != nil {
		return nil, fmt.Errorf("unexpected error when creating the load balancer: %v", err)
	}
	t := &CustomTransport{
		Transport: *http.DefaultTransport.(*http.Transport).Clone(),
		lb:        lb,
	}
	go t.lb.refreshIPs(upstream.DNSRefreshInterval)

	d := &net.Dialer{
		Timeout: dialerTimeout,
	}
	t.DialContext = d.DialContext
	t.TLSHandshakeTimeout = TLSHandshakeTimeout
	t.ResponseHeaderTimeout = responseHeaderTimeout

	return t, nil
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Authorization")

	ctx, cancel := context.WithTimeout(r.Context(), p.upstream.HTTPClientTimeout)
	defer cancel()
	r = r.WithContext(ctx)

	p.reverseProxy.ServeHTTP(w, r)
}

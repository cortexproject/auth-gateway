package gateway

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	DISTRIBUTOR = "distributor"
	FRONTEND    = "frontend"
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
}

type Proxy struct {
	targetURL    *url.URL
	timeouts     Upstream
	reverseProxy *httputil.ReverseProxy
}

func NewProxy(targetURL string, timeouts Upstream, component string) (*Proxy, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	if url.Scheme == "" {
		return nil, fmt.Errorf("invalid URL scheme: %s", targetURL)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(url)
	reverseProxy.Transport = customTransport(component, timeouts)
	originalDirector := reverseProxy.Director
	reverseProxy.Director = customDirector(url, originalDirector)

	if timeouts.HTTPClientTimeout == 0 {
		timeouts.HTTPClientTimeout = defaultTimeoutValues[component].HTTPClientTimeout
	}

	return &Proxy{
		targetURL:    url,
		timeouts:     timeouts,
		reverseProxy: reverseProxy,
	}, nil
}

// it may seem redundant right now but I plan to use this later
func customDirector(targetURL *url.URL, originalDirector func(*http.Request)) func(*http.Request) {
	return func(r *http.Request) {
		originalDirector(r)
	}
}

func customTransport(component string, timeouts Upstream) *http.Transport {
	dialerTimeout := timeouts.HTTPClientDialerTimeout * time.Second
	if dialerTimeout == 0 {
		dialerTimeout = defaultTimeoutValues[component].HTTPClientDialerTimeout
	}
	TLSHandshakeTimeout := timeouts.HTTPClientTLSHandshakeTimeout * time.Second
	if TLSHandshakeTimeout == 0 {
		TLSHandshakeTimeout = defaultTimeoutValues[component].HTTPClientTLSHandshakeTimeout
	}
	responseHeaderTimeout := timeouts.HTTPClientResponseHeaderTimeout * time.Second
	if responseHeaderTimeout == 0 {
		responseHeaderTimeout = defaultTimeoutValues[component].HTTPClientResponseHeaderTimeout
	}

	t := http.DefaultTransport.(*http.Transport).Clone()
	d := &net.Dialer{
		Timeout: dialerTimeout,
	}
	t.DialContext = d.DialContext
	t.TLSHandshakeTimeout = TLSHandshakeTimeout
	t.ResponseHeaderTimeout = responseHeaderTimeout

	return t
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Authorization")

	ctx, cancel := context.WithTimeout(r.Context(), p.timeouts.HTTPClientTimeout)
	defer cancel()
	r = r.WithContext(ctx)

	p.reverseProxy.ServeHTTP(w, r)
}

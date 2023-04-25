package gateway

import (
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
		HTTPClientDialerTimeout:         time.Second * 5,
		HTTPClientTLSHandshakeTimeout:   time.Second * 5,
		HTTPClientResponseHeaderTimeout: time.Second * 5,
	},
	FRONTEND: {
		HTTPClientDialerTimeout:         time.Second * 5,
		HTTPClientTLSHandshakeTimeout:   time.Second * 5,
		HTTPClientResponseHeaderTimeout: time.Second * 5,
	},
}

type Proxy struct {
	targetURL *url.URL

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

	return &Proxy{
		targetURL:    url,
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
	dialerTimeout := timeouts.HTTPClientDialerTimeout
	if dialerTimeout == 0 {
		dialerTimeout = defaultTimeoutValues[component].HTTPClientDialerTimeout
	}
	TLSHandshakeTimeout := timeouts.HTTPClientTLSHandshakeTimeout
	if TLSHandshakeTimeout == 0 {
		TLSHandshakeTimeout = defaultTimeoutValues[component].HTTPClientTLSHandshakeTimeout
	}
	responseHeaderTimeout := timeouts.HTTPClientResponseHeaderTimeout
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
	p.reverseProxy.ServeHTTP(w, r)
}

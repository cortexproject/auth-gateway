package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const (
	DISTRIBUTOR = "distributor"
	FRONTEND    = "frontend"
)

var defaultTimeoutValues map[string]Timeouts = map[string]Timeouts{
	DISTRIBUTOR: {
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 5,
	},
	FRONTEND: {
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 5,
		IdleTimeout:  time.Second * 5,
	},
}

type Proxy struct {
	targetURL *url.URL

	reverseProxy *httputil.ReverseProxy
}

func NewProxy(targetURL string, timeouts Timeouts, component string) (*Proxy, error) {
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

func customTransport(component string, timeouts Timeouts) *http.Transport {
	readTimeout := timeouts.ReadTimeout
	if readTimeout == 0 {
		readTimeout = defaultTimeoutValues[component].ReadTimeout
	}
	writeTimeout := timeouts.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = defaultTimeoutValues[component].WriteTimeout
	}
	idleTimeout := timeouts.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = defaultTimeoutValues[component].IdleTimeout
	}

	return &http.Transport{
		ResponseHeaderTimeout: readTimeout,
		ExpectContinueTimeout: writeTimeout,
		IdleConnTimeout:       idleTimeout,
	}
}

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Authorization")
	p.reverseProxy.ServeHTTP(w, r)
}

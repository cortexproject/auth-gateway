package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	targetName string
	targetURL  *url.URL

	reverseProxy *httputil.ReverseProxy
}

func NewProxy(targetURL string, targetName string) (*Proxy, error) {
	url, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	if url.Scheme == "" {
		return nil, fmt.Errorf("invalid URL scheme: %s", targetURL)
	}

	reverseProxy := httputil.NewSingleHostReverseProxy(url)
	originalDirector := reverseProxy.Director
	reverseProxy.Director = customDirector(url, originalDirector)

	return &Proxy{
		targetName:   targetName,
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

func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	r.Header.Del("Authorization")
	p.reverseProxy.ServeHTTP(w, r)
}

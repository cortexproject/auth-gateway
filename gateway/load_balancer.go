package gateway

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

type DNSResolver interface {
	LookupIP(string) ([]net.IP, error)
}

type DefaultDNSResolver struct{}

func (d DefaultDNSResolver) LookupIP(hostname string) ([]net.IP, error) {
	return net.LookupIP(hostname)
}

type roundRobinLoadBalancer struct {
	hostname     string
	ips          []string
	currentIndex int
	transport    http.RoundTripper
	resolveIPs   func(hostname string) ([]net.IP, error)
	sync.RWMutex
}

func newRoundRobinLoadBalancer(hostname string, resolver func(hostname string) ([]net.IP, error)) *roundRobinLoadBalancer {
	lb := &roundRobinLoadBalancer{
		hostname:   hostname,
		transport:  http.DefaultTransport,
		resolveIPs: resolver,
	}

	// Resolve IPs initially
	ips, err := lb.resolveIPs(hostname)
	if err != nil {
		log.Printf("Failed to resolve IPs for hostname %s: %v", lb.hostname, err)
	} else {
		lb.ips = ipsToStrings(ips)
	}

	return lb
}

func (lb *roundRobinLoadBalancer) roundTrip(req *http.Request) (*http.Response, error) {
	lb.Lock()
	defer lb.Unlock()

	if len(lb.ips) == 0 {
		// TODO: replace format error with a log statement
		return nil, fmt.Errorf("no IP addresses available")
	}

	ip := lb.getNextIP()
	req.URL.Host = strings.Replace(req.URL.Host, lb.hostname, ip, 1)
	lb.currentIndex++

	return lb.transport.RoundTrip(req)
}

func (lb *roundRobinLoadBalancer) getNextIP() string {
	return lb.ips[lb.currentIndex%len(lb.ips)]
}

func (lb *roundRobinLoadBalancer) safeGetNextIP() string {
	lb.RLock()
	defer lb.RUnlock()

	return lb.getNextIP()
}

// Refresh IPs periodically
func (lb *roundRobinLoadBalancer) refreshIPs(refreshInterval time.Duration) {
	for {
		ips, err := lb.resolveIPs(lb.hostname)
		if err != nil {
			// TODO: replace std library log package with logrus
			log.Printf("Failed to resolve IPs for hostname %s: %v", lb.hostname, err)
		} else {
			lb.Lock()
			lb.ips = ipsToStrings(ips)
			lb.currentIndex = 0
			lb.Unlock()
		}
		time.Sleep(refreshInterval)
	}
}

func ipsToStrings(ips []net.IP) []string {
	strs := make([]string, len(ips))
	for i, ip := range ips {
		strs[i] = ip.String()
	}
	return strs
}

// CustomTransport wraps http.Transport and embeds the round-robin load balancer.
type CustomTransport struct {
	http.Transport
	lb *roundRobinLoadBalancer
}

// RoundTrip sends the HTTP request using round-robin load balancing.
func (ct *CustomTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return ct.lb.roundTrip(req)
}

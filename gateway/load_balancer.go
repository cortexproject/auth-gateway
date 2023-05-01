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

type roundRobinLoadBalancer struct {
	hostname     string
	ips          []string
	currentIndex int
	transport    http.RoundTripper
	sync.RWMutex
}

func newRoundRobinLoadBalancer(hostname string) *roundRobinLoadBalancer {
	lb := &roundRobinLoadBalancer{
		hostname:  hostname,
		transport: http.DefaultTransport,
	}

	// Resolve IPs initially
	ips, err := net.LookupIP(lb.hostname)
	if err != nil {
		log.Printf("Failed to resolve IPs for hostname %s: %v", lb.hostname, err)
	} else {
		lb.ips = ipsToStrings(ips)
	}

	return lb
}

func (lb *roundRobinLoadBalancer) roundTrip(req *http.Request) (*http.Response, error) {
	lb.RLock()
	defer lb.RUnlock()

	if len(lb.ips) == 0 {
		// TODO: replace format error with a log statement
		return nil, fmt.Errorf("no IP addresses available")
	}

	ip := lb.ips[lb.currentIndex%len(lb.ips)]
	req.URL.Host = strings.Replace(req.URL.Host, lb.hostname, ip, 1)
	lb.currentIndex++

	return lb.transport.RoundTrip(req)
}

// Refresh IPs periodically
func (lb *roundRobinLoadBalancer) refreshIPs(refreshInterval time.Duration) {
	for {
		ips, err := net.LookupIP(lb.hostname)
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

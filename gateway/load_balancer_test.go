package gateway

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

type mockDNSResolver struct {
	IPs []net.IP
	Err error
}

func (m mockDNSResolver) LookupIP(hostname string) ([]net.IP, error) {
	return m.IPs, m.Err
}

type customRoundTripper struct{}

func (rt customRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(strings.NewReader("Hello, client")),
		Header:     make(http.Header),
	}
	return resp, nil
}

func TestDistribution(t *testing.T) {
	hostname := "example.com"
	testCases := []struct {
		name            string
		IPs             []net.IP
		numReqs         int
		refreshInterval time.Duration
		tolerance       float64
	}{
		{
			name: "4 IPs, 1000 requests, 1 second refresh interval, 10% tolerance",
			IPs: []net.IP{
				net.ParseIP("192.0.0.1"),
				net.ParseIP("192.0.0.2"),
				net.ParseIP("192.0.0.3"),
				net.ParseIP("192.0.0.4"),
			},
			numReqs:         1000,
			refreshInterval: 1 * time.Second,
			tolerance:       0.1,
		},
		{
			name: "4 IPs, 1000 requests, 2 seconds refresh interval, 10% tolerance",
			IPs: []net.IP{
				net.ParseIP("192.0.0.1"),
				net.ParseIP("192.0.0.2"),
				net.ParseIP("192.0.0.3"),
				net.ParseIP("192.0.0.4"),
			},
			numReqs:         1000,
			refreshInterval: 2 * time.Second,
			tolerance:       0.1,
		},
		{
			name: "4 IPs, 1000 requests, 3 seconds refresh interval, 10% tolerance",
			IPs: []net.IP{
				net.ParseIP("192.0.0.1"),
				net.ParseIP("192.0.0.2"),
				net.ParseIP("192.0.0.3"),
				net.ParseIP("192.0.0.4"),
			},
			numReqs:         1000,
			refreshInterval: 3 * time.Second,
			tolerance:       0.1,
		},
		{
			name: "3 IPs, 1000 requests, 2 seconds refresh interval, 10% tolerance",
			IPs: []net.IP{
				net.ParseIP("192.0.0.1"),
				net.ParseIP("192.0.0.2"),
				net.ParseIP("192.0.0.3"),
			},
			numReqs:         1000,
			refreshInterval: 2 * time.Second,
			tolerance:       0.1,
		},
		{
			name: "10 IPs, 1000 requests, 2 seconds refresh interval, 10% tolerance",
			IPs: []net.IP{
				net.ParseIP("192.0.0.1"),
				net.ParseIP("192.0.0.2"),
				net.ParseIP("192.0.0.3"),
				net.ParseIP("192.0.0.4"),
				net.ParseIP("192.0.0.5"),
				net.ParseIP("192.0.0.6"),
				net.ParseIP("192.0.0.7"),
				net.ParseIP("192.0.0.8"),
				net.ParseIP("192.0.0.9"),
				net.ParseIP("192.0.0.10"),
			},
			numReqs:         1000,
			refreshInterval: 2 * time.Second,
			tolerance:       0.1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockResolver := mockDNSResolver{
				IPs: tc.IPs,
				Err: nil,
			}

			lb := newRoundRobinLoadBalancer(hostname, mockResolver.LookupIP)
			lb.transport = customRoundTripper{}

			go lb.refreshIPs(tc.refreshInterval)

			requestCounts := make(map[string]int)
			for i := 0; i < 1000; i++ {
				req := httptest.NewRequest("GET", "http://"+lb.getNextIP(), nil)
				resp, err := lb.roundTrip(req)
				if err == nil {
					addr := req.URL.Host
					ip := strings.Replace(addr, hostname, "", 1)
					requestCounts[ip]++
					resp.Body.Close()
				} else {
					t.Log(err)
				}
				time.Sleep(10 * time.Millisecond)
			}

			expectedCount := tc.numReqs / len(tc.IPs)
			minCount := int(float64(expectedCount) * (1 - tc.tolerance))
			maxCount := int(float64(expectedCount) * (1 + tc.tolerance))

			for ip, count := range requestCounts {
				if count < minCount || count > maxCount {
					t.Errorf("IP %s received %d requests, which is outside the acceptable range (%d-%d)", ip, count, minCount, maxCount)
				}
			}
		})
	}

}

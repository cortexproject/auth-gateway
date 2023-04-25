package gateway

import (
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/cortexproject/auth-gateway/server"
	"github.com/stretchr/testify/assert"
)

func TestNewGateway(t *testing.T) {
	srv, err := server.New(server.Config{})
	if err != nil {
		t.Fatal(err)
	}

	config := Config{
		Distributor: struct {
			URL          string        `yaml:"url"`
			Paths        []string      `yaml:"paths"`
			ReadTimeout  time.Duration `yaml:"read_timeout"`
			WriteTimeout time.Duration `yaml:"write_timeout"`
			IdleTimeout  time.Duration `yaml:"idle_timeout"`
		}{
			URL:   "http://localhost:8000",
			Paths: nil,
		},
		QueryFrontend: struct {
			URL          string        `yaml:"url"`
			Paths        []string      `yaml:"paths"`
			ReadTimeout  time.Duration `yaml:"read_timeout"`
			WriteTimeout time.Duration `yaml:"write_timeout"`
			IdleTimeout  time.Duration `yaml:"idle_timeout"`
		}{
			URL:   "http://localhost:9000",
			Paths: nil,
		},
	}

	gw, err := New(&config, srv)
	if err != nil {
		t.Fatal(err)
	}

	assert.NotNil(t, gw)
}

func TestStartGateway(t *testing.T) {
	InitLogger(os.Stdout)

	distributorServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	frontendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))

	timeouts := Upstream{
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  5 * time.Second,
	}

	testCases := []struct {
		name           string
		authHeader     string
		config         *Config
		paths          []string
		expectedStatus int
		expectedErr    error
	}{
		{
			name: "default routes",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username",
						Password:       "password",
						ID:             "orgid",
					},
				},
				Distributor: Upstream{
					URL:          distributorServer.URL,
					Paths:        nil,
					ReadTimeout:  timeouts.ReadTimeout,
					WriteTimeout: timeouts.WriteTimeout,
					IdleTimeout:  timeouts.IdleTimeout,
				},
				QueryFrontend: Upstream{
					URL:          frontendServer.URL,
					Paths:        nil,
					ReadTimeout:  timeouts.ReadTimeout,
					WriteTimeout: timeouts.WriteTimeout,
					IdleTimeout:  timeouts.IdleTimeout,
				},
			},
			authHeader: "Basic " + base64.StdEncoding.EncodeToString([]byte("username:password")),
			paths: []string{
				"/api/v1/push",
				"/api/prom/push",
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
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "custom routes",
			config: &Config{
				Tenants: []Tenant{
					{
						Authentication: "basic",
						Username:       "username",
						Password:       "password",
						ID:             "orgid",
					},
				},
				Distributor: Upstream{
					URL: distributorServer.URL,
					Paths: []string{
						"/test/distributor",
					},
					ReadTimeout:  timeouts.ReadTimeout,
					WriteTimeout: timeouts.WriteTimeout,
					IdleTimeout:  timeouts.IdleTimeout,
				},
				QueryFrontend: Upstream{
					URL: frontendServer.URL,
					Paths: []string{
						"/test/frontend",
					},
					ReadTimeout:  timeouts.ReadTimeout,
					WriteTimeout: timeouts.WriteTimeout,
					IdleTimeout:  timeouts.IdleTimeout,
				},
			},
			paths: []string{
				"/test/distributor",
				"/test/frontend",
			},
			authHeader:     "Basic " + base64.StdEncoding.EncodeToString([]byte("username:password")),
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found route",
			config: &Config{
				Distributor: Upstream{
					URL: distributorServer.URL,
					Paths: []string{
						"/test/distributor",
					},
				},
				QueryFrontend: Upstream{
					URL: frontendServer.URL,
					Paths: []string{
						"/test/frontend",
					},
					ReadTimeout:  timeouts.ReadTimeout,
					WriteTimeout: timeouts.WriteTimeout,
					IdleTimeout:  timeouts.IdleTimeout,
				},
			},
			paths: []string{
				"/not/found",
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "invalid distributor proxy",
			config:      &Config{},
			expectedErr: errors.New("invalid URL scheme:"),
		},
		{
			name: "invalid frontend proxy",
			config: &Config{
				Distributor: Upstream{
					URL:   distributorServer.URL,
					Paths: []string{},
				},
			},
			expectedErr: errors.New("invalid URL scheme:"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gw, err := createMockGateway("localhost", 8080, 8081, tc.config)
			if tc.expectedErr == nil && err != nil {
				t.Fatalf("Unexpected error when creating the gateway: %v\n", err)
			}
			if tc.expectedErr != nil && err == nil {
				t.Fatal("Expected an error but got none")
			}
			if tc.expectedErr != nil && err != nil {
				if !strings.Contains(err.Error(), tc.expectedErr.Error()) {
					t.Fatalf("Unexpected error. got: %v want:%v\n", err, tc.expectedErr)
				} else {
					return
				}
			}

			authHandler, _ := gw.srv.GetHTTPHandlers()
			mockServer := httptest.NewServer(authHandler)
			defer mockServer.Close()

			gw.Start(tc.config)

			client := mockServer.Client()

			for _, path := range tc.paths {
				req, _ := http.NewRequest("GET", mockServer.URL+path, nil)
				req.Header.Set("Authorization", tc.authHeader)
				resp, err := client.Do(req)
				if err != nil {
					t.Fatal(err)
				}
				defer resp.Body.Close()

				assert.Equal(t, tc.expectedStatus, resp.StatusCode)
			}
			gw.srv.Shutdown()
		})
	}

}

func createMockGateway(addr string, port int, unAuthPort int, config *Config) (*Gateway, error) {
	cfg := server.Config{
		HTTPListenAddr:                addr,
		HTTPListenPort:                port,
		UnAuthorizedHTTPListenAddr:    addr,
		UnAuthorizedHTTPListenPort:    unAuthPort,
		ServerGracefulShutdownTimeout: 2 * time.Second,
	}

	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	gateway, err := New(config, srv)
	if err != nil {
		srv.Shutdown()
		return nil, err
	}

	return gateway, nil
}

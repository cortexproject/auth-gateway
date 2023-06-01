package server

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "invalid address for auth",
			config: Config{
				HTTPListenAddr:                "http://localhost",
				HTTPListenPort:                8080,
				UnAuthorizedHTTPListenAddr:    "localhost",
				UnAuthorizedHTTPListenPort:    1111,
				ServerGracefulShutdownTimeout: time.Second * 5,
				HTTPServerReadTimeout:         time.Second * 10,
				HTTPServerWriteTimeout:        time.Second * 10,
				HTTPServerIdleTimeout:         time.Second * 15,
			},
			wantErr: errors.New("too many colons in address"),
		},
		{
			name: "invalid address for unauth",
			config: Config{
				HTTPListenAddr:                "localhost",
				HTTPListenPort:                8080,
				UnAuthorizedHTTPListenAddr:    "http://localhost",
				UnAuthorizedHTTPListenPort:    8081,
				ServerGracefulShutdownTimeout: time.Second * 5,
				HTTPServerReadTimeout:         time.Second * 10,
				HTTPServerWriteTimeout:        time.Second * 10,
				HTTPServerIdleTimeout:         time.Second * 15,
			},
			wantErr: errors.New("too many colons in address"),
		},
		{
			name: "valid address",
			config: Config{
				HTTPListenAddr:                "localhost",
				HTTPListenPort:                8080,
				UnAuthorizedHTTPListenAddr:    "localhost",
				UnAuthorizedHTTPListenPort:    8081,
				ServerGracefulShutdownTimeout: time.Second * 5,
				HTTPServerReadTimeout:         time.Second * 10,
				HTTPServerWriteTimeout:        time.Second * 10,
				HTTPServerIdleTimeout:         time.Second * 15,
			},
			wantErr: nil,
		},
		{
			name: "custom routers",
			config: Config{
				HTTPListenAddr:                "localhost",
				HTTPListenPort:                8080,
				UnAuthorizedHTTPListenAddr:    "localhost",
				UnAuthorizedHTTPListenPort:    8081,
				ServerGracefulShutdownTimeout: time.Second * 5,
				HTTPServerReadTimeout:         time.Second * 10,
				HTTPServerWriteTimeout:        time.Second * 10,
				HTTPServerIdleTimeout:         time.Second * 15,
				HTTPRouter:                    http.NewServeMux(),
				UnAuthorizedHTTPRouter:        http.NewServeMux(),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, err := New(tc.config)
			if tc.wantErr != nil && err == nil {
				t.Fatalf("Expected an error but got none. expected: %v\n", tc.wantErr)
			}
			if tc.wantErr == nil && err != nil {
				t.Fatalf("Got an unexpected error: %v\n", err)
			}
			if tc.wantErr != nil && err != nil {
				if !strings.Contains(err.Error(), tc.wantErr.Error()) {
					t.Fatalf("expected %v, got %v\n", tc.wantErr, err)
				} else {
					return
				}
			}
			if tc.wantErr == nil {
				defer server.Shutdown()
				if server.authServer.httpServer.Addr != fmt.Sprintf("%s:%d", tc.config.HTTPListenAddr, tc.config.HTTPListenPort) {
					t.Errorf("Expected server address to be %s:%d, but got %s", tc.config.HTTPListenAddr, tc.config.HTTPListenPort, server.authServer.httpServer.Addr)
				}
			}
			server.Shutdown()
		})
	}
}

func TestServer_RegisterTo(t *testing.T) {
	s := Server{
		authServer: &server{
			http: http.NewServeMux(),
		},
		unAuthServer: &server{
			http: http.NewServeMux(),
		},
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.RegisterTo("/test_auth", testHandler, AUTH)
	s.RegisterTo("/test_unauth", testHandler, UNAUTH)

	req := httptest.NewRequest(http.MethodGet, "/test_auth", nil)
	w := httptest.NewRecorder()

	s.authServer.http.ServeHTTP(w, req)
	resp := w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for AUTH server, but got %d", http.StatusOK, resp.StatusCode)
	}

	req = httptest.NewRequest(http.MethodGet, "/test_unauth", nil)
	w = httptest.NewRecorder()

	s.unAuthServer.http.ServeHTTP(w, req)
	resp = w.Result()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for UNAUTH server, but got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestRun(t *testing.T) {
	tests := []struct {
		name    string
		cfg     Config
		wantErr bool
	}{
		{
			name: "both servers run successfully",
			cfg: Config{
				HTTPRouter:                    http.NewServeMux(),
				HTTPListenAddr:                "localhost",
				HTTPListenPort:                8080,
				HTTPMiddleware:                nil,
				UnAuthorizedHTTPRouter:        http.NewServeMux(),
				UnAuthorizedHTTPListenAddr:    "localhost",
				UnAuthorizedHTTPListenPort:    8081,
				UnAuthorizedHTTPMiddleware:    nil,
				ServerGracefulShutdownTimeout: 5 * time.Second,
				HTTPServerReadTimeout:         10 * time.Second,
				HTTPServerWriteTimeout:        10 * time.Second,
				HTTPServerIdleTimeout:         10 * time.Second,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up a wait group and an HTTP handler to signal when the server has started
			var wg sync.WaitGroup
			wg.Add(2)
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				wg.Done()
			})

			authListener, err := net.Listen(DefaultNetwork, fmt.Sprintf("%s:%d", tt.cfg.HTTPListenAddr, tt.cfg.HTTPListenPort))
			assert.NoError(t, err)
			unAuthListener, err := net.Listen(DefaultNetwork, fmt.Sprintf("%s:%d", tt.cfg.UnAuthorizedHTTPListenAddr, tt.cfg.UnAuthorizedHTTPListenPort))
			assert.NoError(t, err)

			tt.cfg.HTTPRouter.Handle("/ready", handler)
			tt.cfg.UnAuthorizedHTTPRouter.Handle("/ready", handler)

			s := &Server{
				cfg: tt.cfg,
				authServer: &server{
					http:         tt.cfg.HTTPRouter,
					httpServer:   &http.Server{Handler: tt.cfg.HTTPRouter},
					httpListener: authListener,
				},
				unAuthServer: &server{
					http:         tt.cfg.UnAuthorizedHTTPRouter,
					httpServer:   &http.Server{Handler: tt.cfg.UnAuthorizedHTTPRouter},
					httpListener: unAuthListener,
				},
			}

			errChan := make(chan error, 1)
			go func() {
				err := s.Run()
				if err == http.ErrServerClosed {
					err = nil
				}
				errChan <- err
			}()

			// Send requests to both servers to check if they are ready
			go http.Get(fmt.Sprintf("http://%s:%d/ready", tt.cfg.HTTPListenAddr, tt.cfg.HTTPListenPort))
			go http.Get(fmt.Sprintf("http://%s:%d/ready", tt.cfg.UnAuthorizedHTTPListenAddr, tt.cfg.UnAuthorizedHTTPListenPort))

			// Wait for both servers to start and handle the requests
			wg.Wait()

			s.authServer.httpServer.Close()
			s.unAuthServer.httpServer.Close()

			// Check for any errors returned by Run
			err = <-errChan
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestReadyHandler(t *testing.T) {
	cfg := Config{
		HTTPListenAddr:                "localhost",
		HTTPListenPort:                1234,
		UnAuthorizedHTTPListenAddr:    "localhost",
		UnAuthorizedHTTPListenPort:    1235,
		HTTPServerReadTimeout:         5 * time.Second,
		HTTPServerWriteTimeout:        5 * time.Second,
		HTTPServerIdleTimeout:         5 * time.Second,
		ServerGracefulShutdownTimeout: 5 * time.Second,
	}

	s, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create a new server: %v", err)
	}

	tests := []struct {
		name           string
		readyState     bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "Not ready",
			readyState:     false,
			expectedStatus: http.StatusServiceUnavailable,
			expectedBody:   "Not ready!",
		},
		{
			name:           "Ready",
			readyState:     true,
			expectedStatus: http.StatusOK,
			expectedBody:   "Ready!",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s.ready = test.readyState
			req, err := http.NewRequest("GET", "/ready", nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(s.readyHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != test.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, test.expectedStatus)
			}

			if rr.Body.String() != test.expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), test.expectedBody)
			}
		})
	}
	s.Shutdown()
}

func TestCheckPortAvailable(t *testing.T) {
	tests := []struct {
		name          string
		listenAddr    string
		listenPort    int
		wantAvailable bool
	}{
		{
			name:          "port available",
			listenAddr:    "localhost",
			listenPort:    8080,
			wantAvailable: true,
		},
		{
			name:          "port unavailable",
			listenAddr:    "localhost",
			listenPort:    1234,
			wantAvailable: false,
		},
	}

	listener, err := net.Listen(DefaultNetwork, fmt.Sprintf("%s:%d", "localhost", 1234))
	if err != nil {
		t.Fatalf("Failed to create a listener: %v", err)
	}
	defer listener.Close()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			available := checkPortAvailable(tt.listenAddr, tt.listenPort, DefaultNetwork)
			if available != tt.wantAvailable {
				t.Errorf("Expected port %d to be available: %v", tt.listenPort, tt.wantAvailable)
			}
		})
	}
}

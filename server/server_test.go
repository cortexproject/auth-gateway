package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	testCases := []struct {
		name    string
		config  Config
		wantErr error
	}{
		{
			name: "invalid address",
			config: Config{
				HTTPListenAddr:                "http://localhost",
				HTTPListenPort:                8080,
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
				ServerGracefulShutdownTimeout: time.Second * 5,
				HTTPServerReadTimeout:         time.Second * 10,
				HTTPServerWriteTimeout:        time.Second * 10,
				HTTPServerIdleTimeout:         time.Second * 15,
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
			defer server.HTTPListener.Close()
			if server.HTTPServer.Addr != fmt.Sprintf("%s:%d", tc.config.HTTPListenAddr, tc.config.HTTPListenPort) {
				t.Errorf("Expected server address to be %s:%d, but got %s", tc.config.HTTPListenAddr, tc.config.HTTPListenPort, server.HTTPServer.Addr)
			}
		})
	}
}

func TestRunAndShutdown(t *testing.T) {
	cfg := Config{
		HTTPListenAddr:                "localhost",
		HTTPListenPort:                8081,
		ServerGracefulShutdownTimeout: time.Second * 5,
		HTTPServerReadTimeout:         time.Second * 10,
		HTTPServerWriteTimeout:        time.Second * 10,
		HTTPServerIdleTimeout:         time.Second * 15,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	shutdownCh := make(chan struct{})
	go func() {
		<-shutdownCh
		server.Shutdown()
	}()

	close(shutdownCh)

	err = server.Run()
	if err != nil {
		t.Fatalf("Failed to run server: %v", err)
	}
}

func TestRouter(t *testing.T) {
	router := http.NewServeMux()
	router.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "Test passed")
	})

	cfg := Config{
		HTTPListenAddr:                "localhost",
		HTTPListenPort:                8082,
		ServerGracefulShutdownTimeout: time.Second * 5,
		HTTPServerReadTimeout:         time.Second * 10,
		HTTPServerWriteTimeout:        time.Second * 10,
		HTTPServerIdleTimeout:         time.Second * 15,
		HTTPRouter:                    router,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	shutdownCh := make(chan struct{})
	go func() {
		err := server.Run()
		if err != nil {
			t.Logf("Server stopped with error: %v", err)
		}
	}()

	go func() {
		<-shutdownCh
		server.Shutdown()
	}()

	client := http.Client{
		Timeout: time.Second * 5,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("http://%s/test", server.HTTPServer.Addr), nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Failed to perform request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, resp.StatusCode)
	}

	close(shutdownCh)
}

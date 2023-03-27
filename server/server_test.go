package server

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	cfg := Config{
		HTTPListenAddr:                "localhost",
		HTTPListenPort:                8080,
		ServerGracefulShutdownTimeout: time.Second * 5,
		HTTPServerReadTimeout:         time.Second * 10,
		HTTPServerWriteTimeout:        time.Second * 10,
		HTTPServerIdleTimeout:         time.Second * 15,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server.HTTPServer.Addr != fmt.Sprintf("%s:%d", cfg.HTTPListenAddr, cfg.HTTPListenPort) {
		t.Errorf("Expected server address to be %s:%d, but got %s", cfg.HTTPListenAddr, cfg.HTTPListenPort, server.HTTPServer.Addr)
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

	go func() {
		time.Sleep(time.Second)
		server.Shutdown()
	}()

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
		Router:                        router,
	}

	server, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		err := server.Run()
		if err != nil {
			t.Logf("Server stopped with error: %v", err)
		}
	}()

	time.Sleep(time.Second)

	defer server.Shutdown()

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
}

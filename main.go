package main

import (
	"net/http"
	"os"

	"github.com/cortexproject/auth-gateway/gateway"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stdout)
	targetURL := "http://localhost:8888"
	proxy, err := gateway.NewProxy(targetURL)
	if err != nil {
		level.Error(logger).Log("msg", err)
		return
	}

	if len(os.Args) < 2 {
		level.Error(logger).Log("msg", "No configuration file is provided")
		os.Exit(1)
	}

	filePath := os.Args[1]
	tenants, err := gateway.GetTenants(filePath)
	if err != nil {
		level.Error(logger).Log("msg", err)
	}

	http.Handle("/", tenants.Authenticate(proxy))
	level.Error(logger).Log("msg", http.ListenAndServe(":8080", nil))
}

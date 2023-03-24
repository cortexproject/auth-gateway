package main

import (
	"os"

	"github.com/cortexproject/auth-gateway/gateway"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stdout)
	if len(os.Args) < 2 {
		level.Error(logger).Log("msg", "No configuration file is provided")
		os.Exit(1)
	}

	filePath := os.Args[1]
	conf, err := gateway.Init(filePath, logger)
	if err != nil {
		level.Error(logger).Log("msg", err)
	}

	// TODO: below will be implemented in the next PR
	// serverConf := server.Config{}
	// server, err := server.New(serverConf)

	gateway, err := gateway.New(conf)
	if err != nil {
		level.Error(logger).Log("msg", "Could not initiate the gateway")
		os.Exit(1)
	}

	gateway.Start(&conf)

	// TODO: below will be implemented in the next PR
	// server.Run()
}

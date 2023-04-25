package main

import (
	"fmt"
	"os"

	"github.com/cortexproject/auth-gateway/gateway"
	"github.com/cortexproject/auth-gateway/middleware"
	"github.com/cortexproject/auth-gateway/server"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("No configuration file is provided")
		os.Exit(1)
	}

	gateway.InitLogger(os.Stdout)

	filePath := os.Args[1]
	conf, err := gateway.Init(filePath)
	gateway.CheckErr("reading the configuration file", err)

	serverConf := server.Config{
		HTTPListenAddr: conf.Server.Address,
		HTTPListenPort: conf.Server.Port,
		HTTPMiddleware: []middleware.Interface{
			gateway.NewAuthentication(&conf),
		},
		HTTPServerReadTimeout:              conf.Server.ReadTimeout,
		HTTPServerWriteTimeout:             conf.Server.WriteTimeout,
		HTTPServerIdleTimeout:              conf.Server.IdleTimeout,
		UnAuthorizedHTTPListenAddr:         conf.Admin.Address,
		UnAuthorizedHTTPListenPort:         conf.Admin.Port,
		UnAuthorizedHTTPServerReadTimeout:  conf.Admin.ReadTimeout,
		UnAuthorizedHTTPServerWriteTimeout: conf.Admin.WriteTimeout,
		UnAuthorizedHTTPServerIdleTimeout:  conf.Admin.IdleTimeout,
	}
	server, err := server.New(serverConf)
	gateway.CheckErr("initializing the server", err)

	defer server.Shutdown()

	gtw, err := gateway.New(&conf, server)
	gateway.CheckErr("initializing the gateway", err)

	gtw.Start(&conf)

	server.Run()
}

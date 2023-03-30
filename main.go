package main

import (
	"fmt"
	"os"

	"github.com/cortexproject/auth-gateway/gateway"
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
	}
	server, err := server.New(serverConf)
	gateway.CheckErr("initializing the server", err)

	defer server.Shutdown()

	gtw, err := gateway.New(&conf, server)
	gateway.CheckErr("initializing the gateway", err)

	gtw.Start(&conf)

	server.Run()
}

package main

import (
	"fmt"
	"net"
	"net/url"
	"os"
	"strconv"

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

	serverAddr, err := url.Parse(conf.ServerAddress)
	gateway.CheckErr("parsing the url", err)

	host, port, err := net.SplitHostPort(serverAddr.Host)
	gateway.CheckErr("splitting the host and the port", err)

	parsedPort, err := strconv.Atoi(port)
	gateway.CheckErr("converting the port number to an int", err)

	serverConf := server.Config{
		HTTPListenAddr: host,
		HTTPListenPort: parsedPort,
	}
	server, err := server.New(serverConf)
	gateway.CheckErr("initializing the server", err)

	defer server.Shutdown()

	gtw, err := gateway.New(conf, server)
	gateway.CheckErr("initializing the gateway", err)

	gtw.Start(&conf)

	server.Run()
}

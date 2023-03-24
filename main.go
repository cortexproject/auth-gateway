package main

import (
	"net"
	"net/url"
	"os"
	"strconv"

	"github.com/go-kit/log"
	"github.com/go-kit/log/level"

	"github.com/cortexproject/auth-gateway/gateway"
	"github.com/cortexproject/auth-gateway/server"
)

func main() {
	logger := log.NewLogfmtLogger(os.Stdout)
	if len(os.Args) < 2 {
		level.Error(logger).Log("msg", "No configuration file is provided")
		os.Exit(1)
	}

	filePath := os.Args[1]
	conf, err := gateway.Init(filePath, logger)
	gateway.CheckErr("reading the configuration file", err, logger)

	serverAddr, err := url.Parse(conf.ServerAddress)
	gateway.CheckErr("parsing the url", err, logger)

	host, port, err := net.SplitHostPort(serverAddr.Host)
	gateway.CheckErr("splitting the host and the port", err, logger)

	parsedPort, err := strconv.Atoi(port)
	gateway.CheckErr("converting the port number to an int", err, logger)

	serverConf := server.Config{
		HTTPListenAddr: host,
		HTTPListenPort: parsedPort,
	}
	server, err := server.New(serverConf)
	gateway.CheckErr("initializing the server", err, logger)

	defer server.Shutdown()

	gtw, err := gateway.New(conf, server)
	gateway.CheckErr("initializing the gateway", err, logger)

	gtw.Start(&conf)

	server.Run()
}

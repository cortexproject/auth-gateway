package main

import (
	"log"
	"net/http"
	"os"

	"github.com/cortexproject/auth-gateway/gateway"
)

func main() {
	targetURL := "http://localhost:8888"
	proxy, err := gateway.NewProxy(targetURL)
	if err != nil {
		log.Fatal(err)
		return
	}

	if len(os.Args) < 2 {
		log.Println("Please provide a file path as an argument")
		os.Exit(1)
	}

	filePath := os.Args[1]
	gateway.InitTenants(filePath)

	http.Handle("/", gateway.Authenticate(proxy))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

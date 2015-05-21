package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krishicks/goulash/handler"
)

const (
	portVar     = "VCAP_APP_PORT"
	defaultPort = "8080"
)

var (
	port    string
	address string
)

func init() {
	if port = os.Getenv(portVar); port == "" {
		port = defaultPort
	}
	address = fmt.Sprintf(":%s", port)
}

func main() {
	h := handler.New()
	if err := http.ListenAndServe(address, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}

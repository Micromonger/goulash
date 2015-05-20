package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/krishicks/goulash/handler"
)

const (
	PortVar = "VCAP_APP_PORT"
)

var (
	port    string
	address string
)

func init() {
	if port = os.Getenv(PortVar); port == "" {
		port = "8080"
	}
	address = fmt.Sprintf(":%s", port)
}

func main() {
	h := handler.New()
	if err := http.ListenAndServe(address, h); err != nil {
		log.Fatal("Failed to start server", err)
	}
}

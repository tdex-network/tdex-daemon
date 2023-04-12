package main

import (
	"log"
	"os"

	"github.com/tdex-network/tdex-daemon/examples"
)

const (
	daemonAddr   = "localhost:9945"
	explorerAddr = "http://localhost:3001"
)

func main() {
	if err := examples.BuyExample(daemonAddr, explorerAddr); err != nil {
		log.Fatal(err)
	}
	os.Exit(0)
}

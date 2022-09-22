package main

import (
	"context"
	"github.com/libp2p/go-libp2p"
	"log"
	"p2p/discov"
)

func main() {
	host, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create h1: %v", err)
	}

	log.Printf("%v", host.ID())
	discov.AddHost(host, "hasenfurz")

	ctx := context.Background()
	<-ctx.Done()
}

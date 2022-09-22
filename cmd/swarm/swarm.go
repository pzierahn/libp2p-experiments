package main

import (
	"context"
	"flag"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
	"p2p/discov"
)

var (
	connect  = flag.String("c", "", "multipart address")
	provider = flag.Bool("p", false, "enable provider")
	consumer = flag.Bool("c", false, "enable provider")
)

func init() {
	flag.Parse()
}

func main() {
	host, err := libp2p.New()
	if err != nil {
		log.Fatalf("Failed to create h1: %v", err)
	}

	log.Printf("%v", host.ID())

	ctx := context.Background()

	if *connect != "" {
		maddr, err := multiaddr.NewMultiaddr(*connect)
		if err != nil {
			log.Fatalln(err)
		}

		info, err := peer.AddrInfoFromP2pAddr(maddr)
		if err != nil {
			log.Fatalln(err)
		}

		log.Printf("connect to %+v", info)

		err = host.Connect(ctx, *info)
		if err != nil {
			log.Fatalln(err)
		}
	}

	if *provider {
		discov.AddHost(host, "hasenfurz")
		host.SetStreamHandler("test", func(stream network.Stream) {
			log.Printf("test: recieved stream %v", stream.ID())
			_, err = stream.Write([]byte("Hello from " + host.ID().String()))
			if err != nil {
				log.Fatalf("Failed to write stream %v: %v", stream.ID(), err)
			}
		})
	}

	if *consumer {
		discov.AddHost(host, "hasenfurz", func(addr peer.AddrInfo) {
			err = host.Connect(ctx, addr)
			if err != nil {
				log.Fatalf("Failed to connect %v: %v", addr, err)
			}

			stream, err := host.NewStream(ctx, addr.ID, "test")

			buf := make([]byte, 1024)
			_, err = stream.Read(buf)
			if err != nil {
				log.Fatalf("Failed to read stream %v: %v", addr, err)
			}

			log.Printf("message=%s", buf)
		})
	}

	<-ctx.Done()
}

package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
	"log"
	"p2p/discov"
)

var (
	connect  = flag.String("connect", "", "connect to multipart address")
	provider = flag.Bool("provider", false, "enable provider")
	consumer = flag.Bool("consumer", false, "enable consumer")
	port     = flag.Int("port", 0, "enable consumer")
)

func init() {
	flag.Parse()
}

func main() {

	var opts []libp2p.Option

	if *port > 0 {
		opts = append(opts, libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port),
			fmt.Sprintf("/ip4/127.0.0.1/udp/%d/quic", port),
			fmt.Sprintf("/ip6/127.0.0.1/tcp/%d", port),
			fmt.Sprintf("/ip6/127.0.0.1/udp/%d/quic", port),
		))
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create h1: %v", err)
	}

	for _, addr := range host.Addrs() {
		log.Printf("%v", addr)
	}

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

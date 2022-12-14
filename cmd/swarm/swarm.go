package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/libp2p/go-libp2p/p2p/protocol/circuitv2/relay"
	"github.com/multiformats/go-multiaddr"
	"log"
)

var (
	pid     = protocol.ID("hello/1.0.0")
	connect = flag.String("connect", "", "connect to multipart address")
	relayS  = flag.Bool("relay", false, "enable relay")
	port    = flag.Int("port", 0, "enable consumer")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Parse()
}

func main() {

	opts := []libp2p.Option{
		libp2p.EnableRelay(),
		libp2p.EnableAutoRelay(),
		libp2p.EnableHolePunching(),
		libp2p.EnableNATService(),
		libp2p.EnableRelayService(),
		libp2p.NATPortMap(),
	}

	if *port > 0 {
		opts = append(opts, libp2p.ListenAddrStrings(
			fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", *port),
			fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", *port),
			fmt.Sprintf("/ip6/0.0.0.0/tcp/%d", *port),
			fmt.Sprintf("/ip6/0.0.0.0/udp/%d/quic", *port),
		))
	}

	host, err := libp2p.New(opts...)
	if err != nil {
		log.Fatalf("Failed to create h1: %v", err)
	}

	if *relayS {
		_, err := relay.New(host)
		if err != nil {
			log.Fatalf("Failed to create relay: %v", err)
		}
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

		stream, err := host.NewStream(ctx, info.ID, pid)
		if err != nil {
			log.Fatalln(err)
		}

		buf := make([]byte, 1024)
		_, err = stream.Read(buf)
		if err != nil {
			log.Fatalf("Failed to read stream: %v", err)
		}

		log.Printf("message=%s", buf)
	} else {
		log.Printf("listen for %s streams...", pid)

		host.SetStreamHandler(pid, func(stream network.Stream) {
			log.Printf("test: recieved stream %v", stream.ID())
			_, err = stream.Write([]byte("Hello from " + host.ID().String()))
			if err != nil {
				log.Fatalf("Failed to write stream %v: %v", stream.ID(), err)
			}
		})
	}

	log.Println("I can be reached at:")
	for _, addr := range host.Addrs() {
		hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", host.ID()))
		log.Println(addr.Encapsulate(hostAddr))
	}

	<-ctx.Done()
}

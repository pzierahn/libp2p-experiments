package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
	"log"
	"os"
	"p2p/adapter"
)

type discoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

type Key struct {
	PrivKey []byte
}

func (key Key) Private() crypto.PrivKey {
	xxx, err := crypto.UnmarshalPrivateKey(key.PrivKey)
	if err != nil {
		log.Fatalln(err)
	}

	return xxx
}

func credentials() (key Key) {

	if byt, err := os.ReadFile("credentials.host.json"); err == nil {
		err := json.Unmarshal(byt, &key)
		if err != nil {
			log.Fatalln(err)
		}
		return
	}

	priv, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,
		-1,
	)
	if err != nil {
		panic(err)
	}

	keyByte, _ := crypto.MarshalPrivateKey(priv)
	key = Key{
		PrivKey: keyByte,
	}

	byt, _ := json.MarshalIndent(key, "", "  ")
	_ = os.WriteFile("credentials.host.json", byt, 0644)

	return
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// 12D3KooWJRgUavDc5czua2bxtz2dxLQBaqQCC6AL5JJdSaHhnciv
	key := credentials()

	host1, err := libp2p.New(
		libp2p.Identity(key.Private()),
		libp2p.ListenAddrStrings(
			"/ip4/127.0.0.1/tcp/9000",
		),
		libp2p.EnableRelay(),
	)
	if err != nil {
		log.Fatalln(err)
	}

	not := &discoveryNotifee{
		PeerChan: make(chan peer.AddrInfo),
	}
	ser := mdns.NewMdnsService(host1, "xxx-meet", not)
	if err := ser.Start(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for addr := range not.PeerChan {
			jb, _ := json.MarshalIndent(addr, "", "  ")
			log.Printf("Discover: addr=%s", jb)
		}
	}()

	for _, addr := range host1.Addrs() {
		// Build host multiaddress
		hostAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/p2p/%s", host1.ID()))
		log.Printf("hostAddr=%v", addr.Encapsulate(hostAddr).String())
	}

	host1.SetStreamHandler("/echo/1.0.0", func(stream network.Stream) {
		log.Println("/echo/1.0.0: new stream", stream.ID())
	})

	server := grpc.NewServer()

	host1.SetStreamHandler("/grpc/1.0.0", func(stream network.Stream) {
		log.Println("/grpc/1.0.0: new stream", stream.ID())

		conn := adapter.NewAdapter(stream)
		if err := server.Serve(conn); err != nil {
			log.Fatalf("stream=%v err=%v", stream.ID(), err)
		}
	})

	log.Println("listening for connections")

	ctx := context.Background()
	<-ctx.Done()
}

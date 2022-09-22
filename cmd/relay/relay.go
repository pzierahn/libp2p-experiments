package main

import (
	"context"
	"encoding/json"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/libp2p/go-libp2p/p2p/net/swarm"
	"log"
	"p2p/discov"
	"time"

	"github.com/libp2p/go-libp2p"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"

	ma "github.com/multiformats/go-multiaddr"
)

func discoverService(name string, ho host.Host) {
	noti := discov.New()
	ser := mdns.NewMdnsService(ho, "xxx-meet", noti)
	if err := ser.Start(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for addr := range noti.PeerChan {
			jb, _ := json.MarshalIndent(addr, "", "  ")
			log.Printf("Discover: name=%s hostID=%v addr=%s", name, ho.ID(), jb)
		}
	}()
}

func main() {
	// Create three libp2p hosts, enable relay client capabilities on all
	// of them.

	// Tell the host use relays
	h1, err := libp2p.New(libp2p.EnableRelay())
	if err != nil {
		log.Printf("Failed to create h1: %v", err)
		return
	}
	discoverService("h1", h1)

	// Tell the host to relay connections for other peers (The ability to *use*
	// a relay vs the ability to *be* a relay)
	relayHost, err := libp2p.New(libp2p.DisableRelay())
	if err != nil {
		log.Printf("Failed to create relayHost: %v", err)
		return
	}
	_, err = relayv1.NewRelay(relayHost)
	if err != nil {
		log.Printf("Failed to instantiate relayHost relay: %v", err)
		return
	}

	// Zero out the listen addresses for the host, so it can only communicate
	// via p2p-circuit for our example
	h3, err := libp2p.New(libp2p.ListenAddrs(), libp2p.EnableRelay())
	if err != nil {
		log.Printf("Failed to create h3: %v", err)
		return
	}
	discoverService("h3", h3)

	h2info := peer.AddrInfo{
		ID:    relayHost.ID(),
		Addrs: relayHost.Addrs(),
	}

	// Connect both h1 and h3 to relayHost, but not to each other
	if err := h1.Connect(context.Background(), h2info); err != nil {
		log.Printf("Failed to connect h1 and relayHost: %v", err)
		return
	}
	if err := h3.Connect(context.Background(), h2info); err != nil {
		log.Printf("Failed to connect h3 and relayHost: %v", err)
		return
	}

	// Now, to test things, let's set up a protocol handler on h3
	h3.SetStreamHandler("/cats", func(s network.Stream) {
		log.Println("Meow! It worked!")
		s.Close()
	})

	_, err = h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err == nil {
		log.Println("Didn't actually expect to get a stream here. What happened?")
		return
	}
	log.Printf("Okay, no connection from h1 to h3: %v", err)
	log.Println("Just as we suspected")

	time.Sleep(time.Second * 5)

	// Creates a relay address to h3 using relayHost as the relay
	relayaddr, err := ma.NewMultiaddr("/p2p/" + relayHost.ID().String() + "/p2p-circuit/ipfs/" + h3.ID().String())
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("relayaddr=%v", relayaddr)

	for _, peerId := range relayHost.Network().Peers() {
		log.Printf("peerId=%v", peerId)
	}

	// Since we just tried and failed to dial, the dialer system will, by default
	// prevent us from redialing again so quickly. Since we know what we're doing, we
	// can use this ugly hack (it's on our TODO list to make it a little cleaner)
	// to tell the dialer "no, its okay, let's try this again"
	h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID())

	h3relayInfo := peer.AddrInfo{
		ID:    h3.ID(),
		Addrs: []ma.Multiaddr{relayaddr},
	}
	if err := h1.Connect(context.Background(), h3relayInfo); err != nil {
		log.Fatalf("Failed to connect h1 and h3: %v", err)
	}

	// Woohoo! we're connected!
	s, err := h1.NewStream(context.Background(), h3.ID(), "/cats")
	if err != nil {
		log.Fatalf("huh, this should have worked: %v", err)
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream
}

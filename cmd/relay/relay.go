package main

import (
	"context"
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"

	"github.com/libp2p/go-libp2p"
	relayv1 "github.com/libp2p/go-libp2p/p2p/protocol/circuitv1/relay"

	ma "github.com/multiformats/go-multiaddr"
)

func main() {
	// Create three libp2p hosts, enable relay client capabilities on all
	// of them.

	// Tell the host use relays
	h1, err := libp2p.New(libp2p.EnableRelay())
	if err != nil {
		log.Printf("Failed to create h1: %v", err)
		return
	}

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
	//h1.Network().(*swarm.Swarm).Backoff().Clear(h3.ID())

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
		log.Fatalf("huh, this should have worked: ", err)
	}

	s.Read(make([]byte, 1)) // block until the handler closes the stream

	log.Printf("\n\n######################### h1")
	for _, addr := range h1.Addrs() {
		// Build host multiaddress
		hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", h1.ID()))
		log.Printf("h1 %v", addr.Encapsulate(hostAddr).String())
	}

	log.Printf("\n\n######################### relayHost")
	for _, addr := range relayHost.Addrs() {
		// Build host multiaddress
		hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", relayHost.ID()))
		log.Printf("relayHost %v", addr.Encapsulate(hostAddr).String())
	}

	log.Printf("\n\n######################### h3")
	for _, addr := range h3.Addrs() {
		// Build host multiaddress
		hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/p2p/%s", h3.ID()))
		log.Printf("h3 %v", addr.Encapsulate(hostAddr).String())
	}
}

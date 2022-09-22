package main

import (
	"context"
	"encoding/json"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/core/peerstore"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"github.com/multiformats/go-multiaddr"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"p2p/adapter"
	"p2p/discov"
	"time"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	hostAddr := "/ip4/127.0.0.1/tcp/9000/p2p/12D3KooWJRgUavDc5czua2bxtz2dxLQBaqQCC6AL5JJdSaHhnciv"

	// Turn the targetPeer into a multiaddr.
	maddr, err := multiaddr.NewMultiaddr(hostAddr)
	if err != nil {
		log.Fatalln(err)
	}

	// Extract the peer ID from the multiaddr.
	info, err := peer.AddrInfoFromP2pAddr(maddr)
	if err != nil {
		log.Fatalln(err)
	}

	client, err := libp2p.New()
	if err != nil {
		log.Fatalln(err)
	}

	noti := discov.New()
	ser := mdns.NewMdnsService(client, "xxx-meet", noti)
	if err := ser.Start(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for addr := range noti.PeerChan {
			jb, _ := json.MarshalIndent(addr, "", "  ")
			log.Printf("Discover: addr=%s", jb)
		}
	}()

	// We have a peer ID and a targetAddr, so we add it to the peerstore
	// so LibP2P knows how to contact it
	client.Peerstore().AddAddrs(info.ID, info.Addrs, peerstore.PermanentAddrTTL)

	log.Println("sender opening stream")

	// make a new stream from host B to host A
	// it should be handled on host A by the handler we set above because
	// we use the same /echo/1.0.0 protocol
	stream, err := client.NewStream(context.Background(), info.ID, "/grpc/1.0.0")
	if err != nil {
		log.Fatalln(err)
	}

	conn, err := grpc.DialContext(
		context.Background(),
		"Hasenfurz-"+info.ID.String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
		grpc.WithContextDialer(adapter.NewDialAdapter(stream)),
	)

	log.Printf("state=%v", conn.GetState())
	//out, err := io.ReadAll(stream)
	//if err != nil {
	//	log.Fatalln(err)
	//}
	//
	//log.Printf("read reply: %q\n", out)

	time.Sleep(time.Second * 3)

	log.Println("done")
}

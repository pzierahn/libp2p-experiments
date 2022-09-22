package discov

import (
	"encoding/json"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"log"
)

type DiscoveryNotifee struct {
	PeerChan chan peer.AddrInfo
}

func (n *DiscoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	n.PeerChan <- pi
}

func New() *DiscoveryNotifee {
	return &DiscoveryNotifee{
		PeerChan: make(chan peer.AddrInfo),
	}
}

func AddHost(ho host.Host, serviceName string, callbacks ...func(addr peer.AddrInfo)) {
	noti := &DiscoveryNotifee{
		PeerChan: make(chan peer.AddrInfo),
	}
	ser := mdns.NewMdnsService(ho, serviceName, noti)
	if err := ser.Start(); err != nil {
		log.Fatalln(err)
	}

	go func() {
		for addr := range noti.PeerChan {
			jb, _ := json.MarshalIndent(addr, "", "  ")
			log.Printf("Discover: hostID=%v addr=%s", ho.ID(), jb)

			for _, callback := range callbacks {
				callback(addr)
			}
		}
	}()
}

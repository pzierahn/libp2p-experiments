package discov

import "github.com/libp2p/go-libp2p/core/peer"

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

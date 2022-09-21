package adapter

import (
	"context"
	"github.com/libp2p/go-libp2p/core/network"
	"net"
)

type DialAdapter func(ctx context.Context, target string) (conn net.Conn, err error)

func NewDialAdapter(stream network.Stream) (adapter DialAdapter) {

	adapter = func(ctx context.Context, target string) (conn net.Conn, err error) {
		return &P2PConn{
			Stream: stream,
		}, nil
	}

	return
}

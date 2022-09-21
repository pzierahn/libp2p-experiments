package adapter

import (
	"fmt"
	"github.com/libp2p/go-libp2p/core/network"
	"log"
	"net"
	"time"
)

type P2PConn struct {
	Stream network.Stream
}

// Read reads data from the connection.
// Read can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetReadDeadline.
func (c P2PConn) Read(b []byte) (n int, err error) {
	return c.Stream.Read(b)
}

// Write writes data to the connection.
// Write can be made to time out and return an Error with Timeout() == true
// after a fixed time limit; see SetDeadline and SetWriteDeadline.
func (c P2PConn) Write(b []byte) (n int, err error) {
	return c.Stream.Write(b)
}

// Close closes the connection.
// Any blocked Read or Write operations will be unblocked and return errors.
func (c P2PConn) Close() error {
	return c.Stream.Close()
}

// LocalAddr returns the local network address.
func (c P2PConn) LocalAddr() net.Addr {
	return &net.IPAddr{}
}

// RemoteAddr returns the remote network address.
func (c P2PConn) RemoteAddr() net.Addr {
	//log.Printf("################# RemoteAddr:")
	return &net.IPAddr{}
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail with a timeout (see type Error) instead of
// blocking. The deadline applies to all future and pending
// I/O, not just the immediately following call to Read or
// Write. After a deadline has been exceeded, the connection
// can be refreshed by setting a deadline in the future.
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (c P2PConn) SetDeadline(t time.Time) error {
	return c.Stream.SetDeadline(t)
}

// SetReadDeadline sets the deadline for future Read calls
// and any currently-blocked Read call.
// A zero value for t means Read will not time out.
func (c P2PConn) SetReadDeadline(t time.Time) error {
	return c.Stream.SetReadDeadline(t)

}

// SetWriteDeadline sets the deadline for future Write calls
// and any currently-blocked Write call.
// Even if write times out, it may return n > 0, indicating that
// some data was successfully written.
// A zero value for t means Write will not time out.
func (c P2PConn) SetWriteDeadline(t time.Time) error {
	return c.Stream.SetWriteDeadline(t)
}

func NewAdapter(stream network.Stream) (lis net.Listener) {

	ch := make(chan bool, 1)
	ch <- true

	lis = &P2PListenerAdapter{
		stream: stream,
		block:  ch,
	}

	return
}

type P2PListenerAdapter struct {
	stream network.Stream
	block  chan bool
}

// Accept waits for and returns the next connection to the listener.
func (adapter P2PListenerAdapter) Accept() (conn net.Conn, err error) {
	if _, ok := <-adapter.block; ok {
		conn = P2PConn{Stream: adapter.stream}
	} else {
		err = fmt.Errorf("connection cloesed")
	}

	return
}

// Close closes the listener.
// Any blocked Accept operations will be unblocked and return errors.
func (adapter P2PListenerAdapter) Close() error {
	log.Printf("####### Close stream %v", adapter.stream.ID())
	defer close(adapter.block)
	return adapter.stream.Close()
}

// Addr returns the listener's network address.
func (adapter P2PListenerAdapter) Addr() net.Addr {
	return &net.TCPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: 9000,
	}
}

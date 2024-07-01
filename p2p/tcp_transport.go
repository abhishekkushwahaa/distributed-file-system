package p2p

import (
	"fmt"
	"net"
)

// TCPPeer represents the remote node over a TCP established connection
type TCPPeer struct{
	conn net.Conn
	
	// if we dial and retrieve a conn => outbound == true
	// if we accept and retrieve a conn => outbound == false
	outbound bool
}

func NewTCPPeer(conn net.Conn, outbound bool) *TCPPeer{
	return &TCPPeer{
		conn: conn,
		outbound: outbound,
	}
} 

func (p *TCPPeer) Close() error {
	return p.conn.Close()
}

type TCPTransportOpts struct {
	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
	OnPeer  func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	listener net.Listener
	rpcch chan RPC
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport{
	return &TCPTransport{
		TCPTransportOpts: opts,
		rpcch: make(chan RPC),
	}
}

// Consume implements the Transport interface, which will return read-only channel for reading incoming messages
func (t *TCPTransport) Consume() <- chan RPC {
	return t.rpcch
}

func (t *TCPTransport) ListenAndAccept() error{
	var err error

	listener, err := net.Listen("tcp", t.ListenAddr)
	t.listener = listener
	if err != nil {
		return err
	}

	go t.startAcceptLoop()
	return nil
}

func (t *TCPTransport) startAcceptLoop(){
	for {
		conn, err := t.listener.Accept()
		if err != nil {
			fmt.Printf("TCP accept error: %s\n", err)
		}
		fmt.Printf("New Incoming connection %+v\n", conn)
		go t.handleConn(conn)
	}
}

func (t *TCPTransport) handleConn(conn net.Conn){
	var err error

	defer func ()  {
		fmt.Printf("Dropping peer connection: %s", err)	
		conn.Close()
	}()

	peer := NewTCPPeer(conn, true)

	if err = t.HandshakeFunc(peer); err != nil{
		return 
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}


	// Read loop
	rpc := RPC{}
	for {
		err = t.Decoder.Decode(conn, &rpc);
		if err != nil {
			return
		}

		rpc.From = conn.RemoteAddr()

		t.rpcch <- rpc
	}
}
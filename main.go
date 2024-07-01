package main

import (
	"fmt"
	"log"

	"github.com/abhishekkushwahaa/distributed-file-system/p2p"
)

func OnPeer(peer p2p.Peer) error {
	peer.Close()
	return nil
}

func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		HandshakeFunc: p2p.NopHAndshakeFunc,
		Decoder: p2p.DefaultDecoder{},
		OnPeer: OnPeer,
	}

	tr := p2p.NewTCPTransport(tcpopts)

	go func ()  {
		for {
			msg := <- tr.Consume()
			fmt.Printf("%+v\n", msg)
		}	
	}()

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select{}
}
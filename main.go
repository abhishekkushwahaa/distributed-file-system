package main

import (
	"log"

	"github.com/abhishekkushwahaa/distributed-file-system/p2p"
)

func main() {
	tcpopts := p2p.TCPTransportOpts{
		ListenAddr: ":3000",
		HandshakeFunc: p2p.NopHAndshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	}

	tr := p2p.NewTCPTransport(tcpopts)

	if err := tr.ListenAndAccept(); err != nil {
		log.Fatal(err)
	}

	select{}
}
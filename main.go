package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"time"

	"github.com/abhishekkushwahaa/distributed-file-system/p2p"
)

func makeServer(listenAddr string, nodes ...string) *FileServer {
	storageRoot := strings.Replace(listenAddr, ":", "_", -1) + "_network"
	tcptransportOpts := p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NopHAndshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	}
	tcpTransport := p2p.NewTCPTransport(tcptransportOpts)

	fileServerOpts := FileServerOpts{
		EncKey:            newEncryptionKey(),
		StorageRoot:       storageRoot,
		PathTransformFunc: CASPathTransformFunc,
		Transport:         tcpTransport,
		BootstrapNodes:    nodes,
	}

	s := NewFileServer(fileServerOpts)
	tcpTransport.OnPeer = s.OnPeer
	return s
}

func main() {
	s1 := makeServer(":3000")
	s2 := makeServer(":4000", ":3000")
	s3 := makeServer(":5000", ":3000", ":4000")

	go func() { log.Fatal(s1.Start()) }()

	time.Sleep(time.Millisecond * 500)

	go func() { log.Fatal(s2.Start()) }()

	time.Sleep(2 * time.Second)

	go func() { s3.Start() }()

	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("picture_%d.png", i)
		data := bytes.NewReader([]byte("Contents of naturePicture.jpg"))
		err := s3.Store(key, data)
		if err != nil {
			log.Fatalf("Failed to store file on s2: %v", err)
		}

		// if err := s3.store.Delete(key); err != nil {
		// 	log.Fatal(err)
		// }

		r, err := s3.Get(key)
		if err != nil {
			log.Fatalf("Failed to get naturePicture.jpg: %v", err)
		}

		b, err := ioutil.ReadAll(r)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(b))
	}
}

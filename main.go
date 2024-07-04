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

	go func() {
		log.Fatal(s1.Start())
	}()

	time.Sleep(2 * time.Second)

	go func() {
		err := s2.Start()
		if err != nil {
			log.Fatalf("Failed to start s2: %v", err)
		}
	}()

	// for i := 0; i < 2; i++ {
	// 	data := bytes.NewReader([]byte("my big data file here!"))
	// 	s2.Store(fmt.Sprintf("myprivatedata_%d", i), data)
	// 	time.Sleep(time.Millisecond * 5)
	// }

	data := bytes.NewReader([]byte("Contents of naturePicture.jpg"))
	if err := s2.Store("naturePicture.jpg", data); err != nil {
		log.Fatalf("Failed to store file on s2: %v", err)
	}

	r, err := s2.Get("naturePicture.jpg")
	if err != nil {
		log.Fatalf("Failed to get naturePicture.jpg: %v", err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))
}

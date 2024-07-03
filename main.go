package main

import (
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

	time.Sleep(4 * time.Second)

	go s2.Start()
	time.Sleep(4 * time.Second)

	// for i := 0; i < 2; i++ {
	// 	data := bytes.NewReader([]byte("my big data file here!"))
	// 	s2.Store(fmt.Sprintf("myprivatedata_%d", i), data)
	// 	time.Sleep(time.Millisecond * 5)
	// }

	r, err := s2.Get("naturePicture.jpg")
	if err != nil {
		log.Fatal(err)
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(b))

	select {}
}

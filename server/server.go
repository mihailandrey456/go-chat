package server

import (
	"andrewka/chat/broadcaster"
	"crypto/tls"
	"fmt"
	"log"
	"net"
)

func Run(port uint, certFile, keyFile string) {
	listener, err := newListener(port, certFile, keyFile)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Сервер прослушивает localhost:%d\n", port)

	b := broadcaster.New()
	go b.Serve()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn, b)
	}
}

func newListener(port uint, certFile, keyFile string) (net.Listener, error) {
	addr := fmt.Sprintf(":%d", port)
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	config := tls.Config{Certificates: []tls.Certificate{cert}}
	return tls.Listen("tcp", addr, &config)
}

package server

import (
	"andrewka/chat/broadcaster"
	"fmt"
	"log"
	"net"
)

func Run(port uint) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
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

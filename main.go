package main

import (
	"andrewka/chat/server"
	"flag"
)

var port = flag.Uint("port", 8000, "listen port")

func main() {
	flag.Parse()
	server.Run(*port)
}

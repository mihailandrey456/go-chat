package main

import (
	"andrewka/chat/server"
	"flag"
)

var port = flag.Uint("port", 8000, "listen port")
var useTLS = flag.Bool("use-tls", false, "use tls")

func main() {
	flag.Parse()
	server.Run(*port, *useTLS)
}

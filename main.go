package main

import (
	"andrewka/chat/server"
	"flag"
	"log"
)

var port = flag.Uint("port", 8000, "Прослушиваемый порт")
var certFile = flag.String("cert-file", "", "Путь до открытого tls ключа")
var keyFile = flag.String("key-file", "", "Путь до закрытог tls ключа")

func main() {
	flag.Parse()
	if len(*certFile) == 0 || len(*keyFile) == 0 {
		log.Fatal("Неверно указаны пути tls ключей")
	}
	server.Run(*port, *certFile, *keyFile)
}

package server

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"

	"andrewka/chat/broadcaster"
	"andrewka/chat/client"
)

// обработать случай паники
func handleConn(conn net.Conn, bc *broadcaster.Broadcaster) {
	defer conn.Close()

	name := getClientName(conn)
	cli := client.New(client.Addr(conn.RemoteAddr().String()), name)

	go clientWriter(conn, cli)

	cli.InMsg <- "Вы " + cli.Fullname()
	bc.Messages <- "\n" + cli.Fullname() + " подключился"
	bc.Entering <- cli

	doneRead := make(chan struct{})
	go clientReader(conn, cli, doneRead)

input:
	for {
		select {
		case <-time.After(5 * time.Minute):
			break input
		case <-doneRead:
			break input
		case msg := <-cli.OutMsg:
			bc.Messages <- "\n" + cli.Fullname() + ": " + msg
		}
	}

	bc.Leaving <- cli
	bc.Messages <- "\n" + cli.Fullname() + " отключился"
	cli.Close()
}

func getClientName(conn net.Conn) string {
	fmt.Fprintf(conn, "Введите свое имя:\n> ")
	input := bufio.NewScanner(conn)
	for {
		input.Scan()
		name := input.Text()
		if len(name) > 0 {
			return name
		}
		fmt.Fprintf(conn, "Некорректное имя\n> ")
	}
}

func clientReader(conn net.Conn, cli *client.Client, doneRead chan<- struct{}) {
	input := bufio.NewScanner(conn)
	for input.Scan() {
		cli.OutMsg <- input.Text()
	}
	close(doneRead)
}

func clientWriter(conn net.Conn, cli *client.Client) {
	for msg := range cli.InMsg {
		fmt.Fprintln(conn, msg)
	}
}

func Run() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Сервер прослушивает localhost:8000")

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

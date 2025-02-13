package server

import (
	"bufio"
	"log"
	"net"
	"time"
	"errors"

	"andrewka/chat/broadcaster"
	"andrewka/chat/client"
	"andrewka/chat/message"
)

// обработать случай паники
func handleConn(conn net.Conn, bc *broadcaster.Broadcaster) {
	defer conn.Close()

	name, err := getClientName(conn)
	if err != nil {
		return
	}
	cli := client.New(client.Addr(conn.RemoteAddr().String()), name)

	go clientWriter(conn, cli)

	cli.InMsg <- message.Msg{
		From: "Server",
		Content: "Вы " + cli.Fullname(),
	}
	bc.Messages <- message.Msg{
		From: "Server",
		Content: cli.Fullname() + " подключился",
	}
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
			bc.Messages <- msg
		}
	}

	bc.Leaving <- cli
	bc.Messages <- message.Msg{
		From: "Server",
		Content: cli.Fullname() + " отключился",
	}
	cli.Close()
}

func getClientName(conn net.Conn) (string, error) {
	msg := message.Msg{
		From: "Server",
		Content: "Введите свое имя",
	}
	j, err := msg.Marshal()
	if err != nil {
		log.Fatal(err)
	}
	conn.Write(j)

	input := bufio.NewScanner(conn)
	for input.Scan() {
		name := input.Text()
		if len(name) > 0 {
			return name, nil
		}

		msg.Content = "Некорректное имя"
		j, err := msg.Marshal()
		if err != nil {
			log.Fatal(err)
		}
		conn.Write(j)
	}
	return "", errors.New("Не введен имя пользователя")
}

func clientReader(conn net.Conn, cli *client.Client, doneRead chan<- struct{}) {
	input := bufio.NewScanner(conn)
	for input.Scan() {
		cli.OutMsg <- message.Msg{
			From: cli.Fullname(),
			Content: input.Text(),
		}
	}
	close(doneRead)
}

func clientWriter(conn net.Conn, cli *client.Client) {
	for msg := range cli.InMsg {
		j, err := msg.Marshal()
		if err != nil {
			log.Println(err)
		} else {
			conn.Write(j)
		}
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

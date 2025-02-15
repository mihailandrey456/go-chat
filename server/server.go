package server

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"time"

	"andrewka/chat/broadcaster"
	"andrewka/chat/client"
	"andrewka/chat/message"
)

func handleConn(conn net.Conn, bc *broadcaster.Broadcaster) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("Внутренняя ошибка: %v\n", p)
		}
	}()
	defer conn.Close()

	done := make(chan struct{}, 2)

	name, err := getClientName(conn)
	if err != nil {
		log.Println(err)
		return
	}
	cli := client.New(client.Addr(conn.RemoteAddr().String()), name)

	go clientWriter(conn, cli, done)

	cli.InMsg <- message.Msg{
		From:    "Server",
		Content: "Вы " + cli.Fullname(),
	}
	bc.Messages <- message.Msg{
		From:    "Server",
		Content: cli.Fullname() + " подключился",
	}
	bc.Entering <- cli

	clientReader(conn, bc, cli, done)

	bc.Leaving <- cli
	bc.Messages <- message.Msg{
		From:    "Server",
		Content: cli.Fullname() + " отключился",
	}
	cli.Close()
}

func getClientName(conn net.Conn) (name string, err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("Внутренняя ошибка: %v\n", p)
		}
	}()

	msg := message.Msg{
		From:    "Server",
		Content: "Введите свое имя",
	}
	j, err := msg.Marshal()
	if err != nil {
		return "", err
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
			return "", err
		}
		conn.Write(j)
	}
	return "", errors.New("Не введено имя пользователя")
}

func clientReader(conn net.Conn, bc *broadcaster.Broadcaster, cli *client.Client, done chan struct{}) {
	go readClientInput(conn, cli, done)

	for {
		// утечка горутинов??
		go func() {
			<-time.After(5 * time.Minute)
			done <- struct{}{}
		}()

		select {
		case <-done:
			return
		case msg := <-cli.OutMsg:
			bc.Messages <- msg
		}
	}
}

func readClientInput(conn net.Conn, cli *client.Client, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
		if p := recover(); p != nil {
			log.Printf("Внутренняя ошибка: %v\n", p)
		}
	}()

	input := bufio.NewScanner(conn)
	for input.Scan() {
		cli.OutMsg <- message.Msg{
			From:    cli.Fullname(),
			Content: input.Text(),
		}
	}
}

func clientWriter(conn net.Conn, cli *client.Client, done chan<- struct{}) {
	defer func() {
		done <- struct{}{}
		if p := recover(); p != nil {
			log.Printf("Внутренняя ошибка: %v\n", p)
		}
	}()

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

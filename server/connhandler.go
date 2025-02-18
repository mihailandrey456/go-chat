package server

import (
	"andrewka/chat/broadcaster"
	"andrewka/chat/client"
	"andrewka/chat/message"
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type connHandler struct {
	conn net.Conn
	bc   *broadcaster.Broadcaster
	done chan struct{}
}

func newConnHandler(conn net.Conn, bc *broadcaster.Broadcaster) *connHandler {
	return &connHandler{
		conn,
		bc,
		make(chan struct{}, 1),
	}
}

func handleConn(conn net.Conn, bc *broadcaster.Broadcaster) {
	defer func() {
		if p := recover(); p != nil {
			log.Printf("Внутренняя ошибка: %v\n", p)
		}
	}()

	h := newConnHandler(conn, bc)
	defer h.conn.Close()

	name, err := h.readClientName()
	if err != nil {
		log.Println(err)
		return
	}
	cli := client.New(client.Addr(h.conn.RemoteAddr().String()), name)

	go h.clientWriter(cli)

	cli.InMsg <- message.Msg{
		From:    "Server",
		Content: "Вы " + cli.Fullname(),
	}
	h.bc.Messages <- message.Msg{
		From:    "Server",
		Content: cli.Fullname() + " подключился",
	}
	h.bc.Entering <- cli

	h.clientReader(cli)

	h.bc.Leaving <- cli
	h.bc.Messages <- message.Msg{
		From:    "Server",
		Content: cli.Fullname() + " отключился",
	}
	cli.Close()
}

func (h *connHandler) readClientName() (name string, err error) {
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
	if _, err = h.conn.Write(j); err != nil {
		return "", err
	}

	input := bufio.NewScanner(h.conn)
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
		if _, err = h.conn.Write(j); err != nil {
			return "", err
		}
	}
	return "", errors.New("Не введено имя пользователя")
}

func (h *connHandler) clientReader(cli *client.Client) {
	go h.readClientInput(cli)

	for {
		select {
		case <-time.After(5 * time.Minute):
			return
		case <-h.done:
			return
		case msg := <-cli.OutMsg:
			h.bc.Messages <- msg
		}
	}
}

func (h *connHandler) readClientInput(cli *client.Client) {
	defer func() {
		h.done <- struct{}{}
		if p := recover(); p != nil {
			log.Printf("Внутренняя ошибка: %v\n", p)
		}
	}()

	input := bufio.NewScanner(h.conn)
	for input.Scan() {
		cli.OutMsg <- message.Msg{
			From:    cli.Fullname(),
			Content: input.Text(),
		}
	}
}

func (h *connHandler) clientWriter(cli *client.Client) {
	defer func() {
		h.done <- struct{}{}
		if p := recover(); p != nil {
			log.Printf("Внутренняя ошибка: %v\n", p)
		}
	}()

	for msg := range cli.InMsg {
		j, err := msg.Marshal()
		if err != nil {
			log.Println(err)
		} else if _, err = h.conn.Write(j); err != nil {
			return
		}
	}
}

package client

import (
	"andrewka/chat/message"
)

type Addr string

type Client struct {
	Addr   Addr
	name   string
	InMsg  chan message.Msg
	OutMsg chan message.Msg
}

func New(addr Addr, name string) *Client {
	return &Client{
		addr,
		name,
		make(chan message.Msg, 1),
		make(chan message.Msg),
	}
}

func (c Client) Fullname() string {
	return c.name + "@" + string(c.Addr)
}

func (c Client) Close() {
	close(c.InMsg)
	close(c.OutMsg)
}

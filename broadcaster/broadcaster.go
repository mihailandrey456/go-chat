package broadcaster

import (
	"andrewka/chat/client"
	"andrewka/chat/message"
	"fmt"
)

type Broadcaster struct {
	Entering chan *client.Client
	Leaving  chan *client.Client
	Messages chan message.Msg
}

func New() *Broadcaster {
	return &Broadcaster{
		make(chan *client.Client),
		make(chan *client.Client),
		make(chan message.Msg),
	}
}

func (b *Broadcaster) Serve() {
	clients := make(map[client.Addr]*client.Client)

	for {
		select {
		case msg := <-b.Messages:
			for _, cli := range clients {
				cli.InMsg <- msg
			}
		case cli := <-b.Entering:
			cli.InMsg <- message.Msg{
				From:    "Server",
				Content: fmt.Sprintf("В сети %d пользователей", len(clients)),
			}
			clients[cli.Addr] = cli
		case cli := <-b.Leaving:
			delete(clients, cli.Addr)
		}
	}
}

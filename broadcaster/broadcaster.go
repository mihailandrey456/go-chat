package broadcaster

import (
	"andrewka/chat/client"
	"andrewka/chat/message"
	"container/list"
	"fmt"
)

type Broadcaster struct {
	Entering chan *client.Client
	Leaving  chan *client.Client
	Messages chan message.Msg
	history  *list.List
}

func New() *Broadcaster {
	return &Broadcaster{
		make(chan *client.Client),
		make(chan *client.Client),
		make(chan message.Msg),
		list.New(),
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

			if b.history.Len() >= message.HistorySize {
				b.history.Remove(b.history.Front())
			}
			b.history.PushBack(msg)

		case cli := <-b.Entering:
			cli.InMsg <- message.Msg{
				From:    "Server",
				Content: fmt.Sprintf("В сети %d пользователей", len(clients)),
			}
			clients[cli.Addr] = cli

			for e := b.history.Front(); e != nil; e = e.Next() {
				msg, ok := e.Value.(message.Msg)
				if !ok {
					panic("Неожиданный тип элемента в Broadcaster.history")
				}
				cli.InMsg <- msg
			}

		case cli := <-b.Leaving:
			delete(clients, cli.Addr)
		}
	}
}

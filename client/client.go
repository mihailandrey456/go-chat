package client

type Addr string

type Client struct {
	Addr   Addr
	name   string
	InMsg  chan string
	OutMsg chan string
}

func New(addr Addr, name string) *Client {
	return &Client{
		addr,
		name,
		make(chan string),
		make(chan string),
	}
}

func (c Client) Fullname() string {
	return c.name + "@" + string(c.Addr)
}

func (c Client) Close() {
	close(c.InMsg)
	close(c.OutMsg)
}

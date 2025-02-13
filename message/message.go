package message

import (
	"encoding/json"
)

type Msg struct {
	From    string
	Content string
}

func (m Msg) String() string {
	return m.From + ": " + m.Content
}

func (m Msg) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

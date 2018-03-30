package messages

import "encoding/json"

type ArborMessageType uint8

const (
	WELCOME     = 0
	QUERY       = 1
	NEW_MESSAGE = 2
)

type ArborMessage struct {
	Type   ArborMessageType
	Root   string
	Recent []string
	Major  uint8
	Minor  uint8
	*Message
}

func (m *ArborMessage) String() string {

	data, _ := json.Marshal(m)
	dataString := string(data)
	return dataString

}

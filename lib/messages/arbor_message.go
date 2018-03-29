package messages

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

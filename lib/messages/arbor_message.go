package messages

type ArborMessageType uint8

const (
	QUERY          = 0
	NEW_MESSAGE    = 1
	CREATE_MESSAGE = 2
)

type ArborMessage struct {
	Type ArborMessageType
	*Message
}

package arbor

import "encoding/json"

const (
	// WelcomeType should be used as the `Type` field of a WELCOME ProtocolMessage
	WelcomeType = 0
	// QueryType should be used as the `Type` field of a QUERY ProtocolMessage
	QueryType = 1
	// NewMessageType should be used as the `Type` field of a NEW_MESSAGE ProtocolMessage
	NewMessageType = 2
)

// ProtocolMessage represents a message in the Arbor chat protocol. This may or
// may not contain a chat message sent between users.
type ProtocolMessage struct {
	// The type of the message, should be one of the constants defined in this
	// package.
	Type uint8
	// Root is only used in WELCOME messages and identifies the root of this server's message tree
	Root string
	// Recent is only used in WELCOME messages and provides a list of recently-sent message ids
	Recent []string
	// Major is only used in WELCOME messages and identifies the major version number of the protocol version in use
	Major uint8
	// Minor is only used in WELCOME messages and identifies the minor version number of the protocol version in use
	Minor uint8
	// Message is the actual chat message content, if any. This is currently only
	// used in NEW_MESSAGE messages
	*ChatMessage
}

// String returns a JSON representation of the message as a string.
func (m *ProtocolMessage) String() string {

	data, _ := json.Marshal(m)
	dataString := string(data)
	return dataString

}

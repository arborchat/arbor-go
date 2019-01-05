package arbor

import (
	"encoding/json"
	"fmt"
)

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
	// Root is only used in WELCOME messages and identifies the root of this server's message tree
	Root string
	// Recent is only used in WELCOME messages and provides a list of recently-sent message ids
	Recent []string
	// The type of the message, should be one of the constants defined in this
	// package.
	Type uint8
	// Major is only used in WELCOME messages and identifies the major version number of the protocol version in use
	Major uint8
	// Minor is only used in WELCOME messages and identifies the minor version number of the protocol version in use
	Minor uint8
	// Message is the actual chat message content, if any. This is currently only
	// used in NEW_MESSAGE messages
	*ChatMessage
}

// Equals returns true if other is equivalent to the message (has the same data or is the same message)
func (m *ProtocolMessage) Equals(other *ProtocolMessage) bool {
	if (m == nil) != (other == nil) {
		// one is nil and the other is not
		return false
	}
	if m == other {
		// either both nil or pointers to the same address
		return true
	}
	if m.Type != other.Type || m.Root != other.Root || m.Major != other.Major || m.Minor != other.Minor {
		return false
	}
	if !m.ChatMessage.Equals(other.ChatMessage) {
		return false
	}
	if !sameSlice(m.Recent, other.Recent) {
		return false
	}
	return true
}

// Source: https://stackoverflow.com/a/15312097
func sameSlice(a, b []string) bool {
	if (a == nil) != (b == nil) {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// MarshalJSON transforms a ProtocolMessage into JSON
func (m *ProtocolMessage) MarshalJSON() ([]byte, error) {
	switch m.Type {
	case WelcomeType:
		return json.Marshal(struct {
			Root   string
			Recent []string
			Type   uint8
			Major  uint8
			Minor  uint8
		}{Type: m.Type, Root: m.Root, Recent: m.Recent, Major: m.Major, Minor: m.Minor})
	case QueryType:
		return json.Marshal(struct {
			UUID string
			Type uint8
		}{UUID: m.UUID, Type: m.Type})
	case NewMessageType:
		return json.Marshal(struct {
			*ChatMessage
			Type uint8
		}{ChatMessage: m.ChatMessage, Type: m.Type})
	default:
		return nil, fmt.Errorf("Unknown message type, could not marshal")
	}
}

// String returns a JSON representation of the message as a string.
func (m *ProtocolMessage) String() string {
	data, _ := json.Marshal(m) // nolint: gosec
	dataString := string(data)
	return dataString
}

// IsValid returns whether the message has the minimum correct fields for its message
// type.
func (m *ProtocolMessage) IsValid() bool {
	switch m.Type {
	case WelcomeType:
		return m.IsValidWelcome()
	case QueryType:
		return m.IsValidQuery()
	case NewMessageType:
		return m.IsValidNew()
	default:
		return false
	}
}

// IsValidWelcome checks that the message is a valid Welcome message.
func (m *ProtocolMessage) IsValidWelcome() bool {
	switch {
	case m.Type != WelcomeType:
		fallthrough
	case m.Major == 0 && m.Minor == 0:
		fallthrough
	case m.Recent == nil:
		fallthrough
	case m.Root == "":
		return false
	}
	return true
}

// IsValidNew checks that the message is a valid New message.
func (m *ProtocolMessage) IsValidNew() bool {
	switch {
	case m.Type != NewMessageType:
		fallthrough
	case m.ChatMessage == nil:
		fallthrough
	case m.Username == "":
		fallthrough
	case m.Content == "":
		fallthrough
	case m.Timestamp == 0:
		return false
	}
	return true
}

// IsValidQuery checks that the message is a valid Query message.
func (m *ProtocolMessage) IsValidQuery() bool {
	switch {
	case m.Type != QueryType:
		fallthrough
	case m.ChatMessage == nil:
		fallthrough
	case m.UUID == "":
		return false
	}
	return true
}

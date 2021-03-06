package arbor

import (
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
)

// Post is a single chat message within Arbor
type Post ChatMessage

// ChatMessage represents a single chat message sent between users.
type ChatMessage struct {
	UUID      string
	Parent    string
	Content   string
	Username  string
	Timestamp int64
}

// NewChatMessage constructs a ChatMessage with the provided content.
// It's not necessary to create messages with this function,
// but it sets the timestamp for you.
func NewChatMessage(content string) (*ChatMessage, error) {
	return &ChatMessage{
		Parent:    "",
		Content:   content,
		Timestamp: time.Now().Unix(),
	}, nil

}

// AssignID generates a new UUID and sets it as the ID for the
// message.
func (m *ChatMessage) AssignID() error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrapf(err, "Unable to generate UUID")
	}
	m.UUID = id.String()
	return nil
}

// Reply returns a new message with the given content that has
// its parent, content, and timestamp already configured.
func (m *ChatMessage) Reply(content string) (*ChatMessage, error) {
	reply, err := NewChatMessage(content)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to reply")
	}
	reply.Parent = m.UUID
	return reply, nil
}

// Equals compares all message fields to determine whether two messages
// are the same.
func (m *ChatMessage) Equals(other *ChatMessage) bool {
	if (m == nil) != (other == nil) {
		// one is nil and the other is not
		return false
	}
	if m == other {
		// either both nil or pointers to the same address
		return true
	}
	return m.UUID == other.UUID && m.Parent == other.Parent && m.Content == other.Content && m.Username == other.Username && m.Timestamp == other.Timestamp
}

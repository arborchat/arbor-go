package messages

import (
	"time"

	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
)

// Message represents a single chat message sent between users.
type Message struct {
	UUID      string
	Parent    string
	Content   string
	Username  string
	Timestamp int64
}

// NewMessage constructs a Message with the provided content.
// It's not necessary to create messages with this function,
// but it sets the timestamp for you.
func NewMessage(content string) (*Message, error) {
	return &Message{
		Parent:    "",
		Content:   content,
		Timestamp: time.Now().Unix(),
	}, nil

}

// AssignID generates a new UUID and sets it as the ID for the
// message.
func (m *Message) AssignID() error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrapf(err, "Unable to generate UUID")
	}
	m.UUID = id.String()
	return nil
}

// Reply returns a new message with the given content that has
// its parent, content, and timestamp already configured.
func (m *Message) Reply(content string) (*Message, error) {
	reply, err := NewMessage(content)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to reply")
	}
	reply.Parent = m.UUID
	return reply, nil
}

package messages

import (
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"time"
)

type Message struct {
	UUID      string
	Parent    string
	Content   string
	Username  string
	Timestamp int64
}

func NewMessage(content string) (*Message, error) {
	return &Message{
		Parent:    "",
		Content:   content,
		Timestamp: time.Now().Unix(),
	}, nil

}

func (m *Message) AssignID() error {
	id, err := uuid.NewV4()
	if err != nil {
		return errors.Wrapf(err, "Unable to generate UUID")
	}
	m.UUID = id.String()
	return nil
}

func (m *Message) Reply(content string) (*Message, error) {
	reply, err := NewMessage(content)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to reply")
	}
	reply.Parent = m.UUID
	return reply, nil
}

package messages

import (
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
)

type Message struct {
	UUID    string
	Parent  string
	Content string
}

func NewMessage(content string) (*Message, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to generate UUID")
	}
	return &Message{
		UUID:    id.String(),
		Parent:  "",
		Content: content,
	}, nil

}

func (m *Message) Reply(content string) (*Message, error) {
	reply, err := NewMessage(content)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to reply")
	}
	reply.Parent = m.UUID
	return reply, nil
}

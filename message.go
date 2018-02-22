package main

import (
	uuid "github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
)

type Message struct {
	*uuid.UUID
	parent  *uuid.UUID
	Content string
}

func NewMessage(content string) (*Message, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to generate UUID")
	}
	return &Message{
		UUID:    id,
		parent:  nil,
		Content: content,
	}, nil

}

func (m *Message) Reply(content string) (*Message, error) {
	reply, err := NewMessage(content)
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to reply")
	}
	reply.parent = m.UUID
	return reply, nil
}

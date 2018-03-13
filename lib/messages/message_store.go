package messages

import (
	"sync"
)

type Store struct {
	sync.RWMutex
	m map[string]*Message
}

func NewStore() *Store {
	return &Store{
		m: make(map[string]*Message),
	}
}

func (s *Store) Get(uuid string) *Message {
	s.RLock()
	msg, _ := s.m[uuid]
	s.RUnlock()
	return msg
}

func (s *Store) Add(msg *Message) {
	s.Lock()
	s.m[msg.UUID] = msg
	s.Unlock()
}

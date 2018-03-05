package main

import (
	"sync"

	"github.com/whereswaldon/arbor/messages"
)

type Tree struct {
	*messages.Store
	sync.RWMutex
	// ChildrenMap is a map from a message's UUID to a slide of UUIDs for each
	// child message of that message
	ChildrenMap map[string][]string
}

func NewTree(s *messages.Store) *Tree {
	return &Tree{Store: s, ChildrenMap: make(map[string][]string)}
}

// Add stores the message and its relationship with its parent within the message
// tree.
func (t *Tree) Add(msg *messages.Message) {
	t.Store.Add(msg)
	t.Lock()
	defer t.Unlock()
	children, ok := t.ChildrenMap[msg.Parent]
	if !ok {
		t.ChildrenMap[msg.Parent] = []string{msg.UUID}
	} else {
		found := false
		for _, m := range children {
			if m == msg.UUID {
				found = true
				break
			}
		}
		if !found {
			t.ChildrenMap[msg.Parent] = append(t.ChildrenMap[msg.Parent], msg.UUID)
		}
	}
}

// Children returns a slice of known child message ids for a given parent message id
func (t *Tree) Children(id string) []string {
	t.RLock()
	defer t.RUnlock()
	children, ok := t.ChildrenMap[id]
	if !ok {
		return []string{}
	}
	return children
}

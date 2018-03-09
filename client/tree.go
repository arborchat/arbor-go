package main

import (
	"log"
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
	if msg.UUID == "" {
		log.Printf("Asked to add message with empty id: %v\n", msg)
	}
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

// getItems returns a slice of messages starting from the current
// leaf message id and working backward along its ancestry. It will never return
// more than maxLength messages in the slice. If it encounters a message ID that is
// unknown, it will return that in the query value. Otherwise, query will return the
// empty string.
func (t *Tree) GetItems(leafId string, maxLength int) (items []*messages.Message, query string) {
	items = make([]*messages.Message, maxLength)
	current := t.Get(leafId)
	if current == nil {
		return items[:0], ""
	}
	count := 1
	parent := ""
	for i := range items {
		items[i] = current
		if current.Parent == "" {
			break
		}
		parent = current.Parent
		current = t.Get(current.Parent)
		if current == nil {
			//request the message corresponding to parentID
			query = parent
			break
		}
		count++
	}
	return items[:min(count, len(items))], query
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
	messages "github.com/whereswaldon/arbor/messages"
)

type MessageListView struct {
	*Tree
	LeafID string
	Query  chan<- string
}

// NewList creates a new MessageListView that uses the provided Tree
// to manage message history. This MessageListView acts as a layout manager
// for the gocui layout package. The method returns both a MessageListView
// and a readonly channel of queries. These queries are message UUIDs that
// the local store has requested the message contents for.
func NewList(store *Tree) (*MessageListView, <-chan string) {
	queryChan := make(chan string)
	return &MessageListView{store, "", queryChan}, queryChan
}

// UpdateMessage sets the provided UUID as the ID of the current "leaf"
// message within the view of the conversation *if* it is a child of
// the previous current "leaf" message.
func (m *MessageListView) UpdateMessage(id string) {
	msg := m.Tree.Get(id)
	if msg.Parent == m.LeafID || m.LeafID == "" {
		m.LeafID = msg.UUID
	}
	m.getItems()
}

// getItems returns a slice of messages starting from the current
// leaf message and working backward along its ancestry.
func (m *MessageListView) getItems() []*messages.Message {
	const length = 100
	items := make([]*messages.Message, length)
	current := m.Tree.Get(m.LeafID)
	if current == nil {
		return items[:0]
	}
	count := 1
	parent := ""
	for i := range items {
		items[i] = current
		if current.Parent == "" {
			break
		}
		parent = current.Parent
		current = m.Tree.Get(current.Parent)
		if current == nil {
			//request the message corresponding to parentID
			m.Query <- parent
			break
		}
		count++
	}
	return items[:min(count, len(items))]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Layout builds a message history in the provided UI
func (m *MessageListView) Layout(ui *gocui.Gui) error {
	maxX, maxY := ui.Size()

	// determine input box coordinates
	inputUX := 0
	inputUY := maxY - 4
	inputW := maxX - 1
	inputH := 3
	if v, err := ui.SetView("message-input", inputUX, inputUY, inputUX+inputW, inputUY+inputH); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println(err)
			return err
		}
		v.Title = "Compose"
		v.Editable = true
		v.Wrap = true
	}
	items := m.getItems()
	currentY := inputUY - 1
	height := 2
	for i, item := range items {
		if currentY < 4 {
			break
		}
		log.Printf("using view coordinates (%d,%d) to (%d,%d)\n",
			0, currentY-height, maxX-1, currentY)
		log.Printf("creating view: %d", i)
		view, err := ui.SetView(fmt.Sprintf("%d", i), 0, currentY-height, maxX-1, currentY)
		if err != nil {
			if err != gocui.ErrUnknownView {
				log.Panicln("unable to create view for message: ", err)
			}
		}
		view.Clear()
		siblings := m.Children(item.Parent)
		fmt.Fprintf(view, "%d siblings| ", len(siblings)-1)
		fmt.Fprint(view, item.Content)
		currentY -= height + 1
	}
	return nil
}

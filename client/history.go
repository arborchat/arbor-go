package main

import (
	"fmt"
	"log"

	"github.com/jroimartin/gocui"
	messages "github.com/whereswaldon/arbor/messages"
)

type MessageListView struct {
	*messages.Store
	LeafID string
	Query  chan string
}

func NewList(store *messages.Store, queryChan chan string) *MessageListView {
	return &MessageListView{store, "", queryChan}
}

func (m *MessageListView) UpdateMessage(id string) {
	msg := m.Store.Get(id)
	if msg.Parent == m.LeafID || m.LeafID == "" {
		m.LeafID = msg.UUID
	}
	m.getItems()
}

func (m *MessageListView) getItems() []string {
	const length = 100
	items := make([]string, length)
	current := m.Store.Get(m.LeafID)
	if current == nil {
		return items
	}
	parentID := current.Parent
	for i := range items {
		if parentID == "" {
			break
		}
		items[i] = current.Content
		current = m.Store.Get(parentID)
		if current == nil {
			//request the message corresponding to parentID
			m.Query <- parentID
			break
		}
		parentID = current.Parent
	}
	return items
}

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
		fmt.Fprint(view, item)
		currentY -= height + 1
	}
	return nil
}

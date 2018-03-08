package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/jroimartin/gocui"
	"github.com/whereswaldon/arbor/messages"
)

type ThreadView struct {
	Thread   []*messages.Message
	CursorID string
	ViewIDs  map[string]struct{}
	LeafID   string
	sync.RWMutex
}

type History struct {
	*Tree
	ThreadView
	Query chan<- string
}

// NewList creates a new History that uses the provided Tree
// to manage message history. This History acts as a layout manager
// for the gocui layout package. The method returns both a History
// and a readonly channel of queries. These queries are message UUIDs that
// the local store has requested the message contents for.
func NewList(store *Tree) (*History, <-chan string) {
	queryChan := make(chan string)
	return &History{
		Tree: store,
		ThreadView: ThreadView{
			LeafID:   "",
			CursorID: "",
			ViewIDs:  make(map[string]struct{}),
		},
		Query: queryChan,
	}, queryChan
}

// UpdateLeaf sets the provided UUID as the ID of the current "leaf"
// message within the view of the conversation *if* it is a child of
// the previous current "leaf" message. If there is no cursor, the new
// leaf will be set as the cursor.
func (m *History) UpdateLeaf(id string) {
	msg := m.Tree.Get(id)
	if msg.Parent == m.LeafID || m.LeafID == "" {
		m.ThreadView.Lock()
		m.LeafID = msg.UUID
		m.ThreadView.Unlock()
	}
	if m.CursorID == "" {
		m.ThreadView.Lock()
		m.CursorID = msg.UUID
		m.ThreadView.Unlock()
	}
}

func (h *History) destroyOldViews(ui *gocui.Gui) {
	h.ThreadView.Lock()
	// destroy old views
	for id := range h.ViewIDs {
		ui.DeleteView(id)
	}
	// reset ids
	h.ViewIDs = make(map[string]struct{})
	h.ThreadView.Unlock()

}

func (h *History) refreshThread() []*messages.Message {
	h.ThreadView.RLock()
	items, query := h.Tree.GetItems(h.ThreadView.LeafID, 100)
	h.ThreadView.RUnlock()
	h.ThreadView.Lock()
	h.ThreadView.Thread = items // save the computed ancestry of the current thread
	h.ThreadView.Unlock()
	if query != "" {
		h.Query <- query // query for any unknown message in the ancestry
	}
	return items
}

// Layout builds a message history in the provided UI
func (m *History) Layout(ui *gocui.Gui) error {
	m.destroyOldViews(ui)

	maxX, maxY := ui.Size()

	// determine input box coordinates
	inputUX := 0
	inputUY := maxY - 4
	inputW := maxX - 1
	inputH := 3
	if err := m.drawInputView(inputUX, inputUY, inputW, inputH, ui); err != nil {
		return err
	}

	// get the latest history
	items := m.refreshThread()
	currentY := inputUY - 1
	height := 2
	m.ThreadView.Lock()
	defer m.ThreadView.Unlock()
	for _, item := range items {
		if currentY < 4 {
			break
		}
		log.Printf("using view coordinates (%d,%d) to (%d,%d)\n",
			0, currentY-height, maxX-1, currentY)
		log.Printf("creating view: %s", item.UUID)
		view, err := ui.SetView(item.UUID, 0, currentY-height, maxX-1, currentY)
		if err != nil {
			if err != gocui.ErrUnknownView {
				log.Panicln("unable to create view for message: ", err)
			}
		}
		m.ViewIDs[item.UUID] = struct{}{}
		view.Clear()
		siblings := m.Children(item.Parent)
		fmt.Fprintf(view, "%d siblings| ", len(siblings)-1)
		fmt.Fprint(view, item.Content)
		currentY -= height + 1
	}
	if _, ok := m.ViewIDs[m.CursorID]; ok {
		// view with cursor still on screen
		_, err := ui.SetCurrentView(m.CursorID)
		if err != nil {
			return err
		}
	} else if m.LeafID != "" {
		// view with cursor off screen
		m.CursorID = m.LeafID
		_, err := ui.SetCurrentView(m.LeafID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (his *History) drawInputView(x, y, w, h int, ui *gocui.Gui) error {
	if v, err := ui.SetView("message-input", x, y, x+w, y+h); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println(err)
			return err
		}
		v.Title = "Compose"
		v.Editable = true
		v.Wrap = true
	}
	return nil
}

func (m *History) CursorUp(g *gocui.Gui, v *gocui.View) error {
	m.ThreadView.Lock()
	defer m.ThreadView.Unlock()
	if m.CursorID == "" {
		m.CursorID = m.LeafID
	}
	msg := m.Get(m.CursorID)
	if msg == nil {
		m.CursorID = m.LeafID
		log.Println("Error fetching cursor message: %s", m.CursorID)
		return nil
	} else if msg.Parent == "" {
		m.CursorID = m.LeafID
		log.Println("Cannot move cursor up, nil parent for message: %v", msg)
		return nil
	} else if _, ok := m.ViewIDs[msg.Parent]; !ok {
		m.CursorID = m.LeafID
		log.Println("Cannot select parent message, not on screen")
		return nil
	} else {
		m.CursorID = msg.Parent
		_, err := g.SetCurrentView(m.CursorID)
		return err
	}
}

func (m *History) CursorDown(g *gocui.Gui, v *gocui.View) error {
	m.ThreadView.Lock()
	defer m.ThreadView.Unlock()
	if m.CursorID == "" {
		m.CursorID = m.LeafID
	}
	msg := m.Get(m.CursorID)
	if msg == nil {
		m.CursorID = m.LeafID
		log.Println("Error fetching cursor message: %s", m.CursorID)
		return nil
	}
	// get the children of the cursor message
	children := m.Children(m.CursorID)
	if len(children) == 0 {
		log.Println("Cannot move cursor down, no children for message: %v", msg)
		return nil
	}
	// find the child that is visible
	onscreen := ""
	for _, child := range children {
		if _, ok := m.ViewIDs[child]; ok {
			onscreen = child
		}
	}

	// if no children are visible
	if onscreen == "" {
		m.CursorID = m.LeafID
		log.Println("Cannot select child message, none on screen")
		return nil
	} else { // select the visible child
		m.CursorID = onscreen
		_, err := g.SetCurrentView(m.CursorID)
		return err
	}
}

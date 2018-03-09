package main

import (
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/jroimartin/gocui"
	wrap "github.com/mitchellh/go-wordwrap"
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
	m.ThreadView.Lock()
	if msg.Parent == m.LeafID || m.LeafID == "" {
		m.LeafID = msg.UUID
	}
	if m.CursorID == "" {
		m.CursorID = msg.UUID
	}
	m.ThreadView.Unlock()
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
	items, query := h.Tree.GetItems(h.ThreadView.LeafID, 1024)
	h.ThreadView.RUnlock()
	h.ThreadView.Lock()
	h.ThreadView.Thread = items // save the computed ancestry of the current thread
	h.ThreadView.Unlock()
	if query != "" {
		log.Println("Querying for message: ", query)
		h.Query <- query // query for any unknown message in the ancestry
	}
	return items
}

func (h *History) Cursor() string {
	h.ThreadView.RLock()
	defer h.ThreadView.RUnlock()
	return h.ThreadView.CursorID
}

// Layout builds a message history in the provided UI
func (m *History) Layout(ui *gocui.Gui) error {
	m.destroyOldViews(ui)

	maxX, maxY := ui.Size()

	//TODO: draw input box iff we are replying to a message

	// get the latest history
	thread := m.refreshThread()
	totalY := maxY // how much vertical space is left for drawing messages

	cursorY := (totalY - 2) / 2
	cursorX := 0
	cursorId := m.Cursor()
	if cursorId == "" {
		return nil
	}
	err, cursorHeight := m.drawView(cursorX, cursorY, maxX-1, down, true, cursorId, ui) //draw the cursor message
	if err != nil {
		log.Println("error drawing cursor view: ", err)
		return err
	}

	var currentIdxBelow int = -1
	var currentIdxAbove int = -1
	for i, message := range thread {
		if message.UUID == cursorId {
			currentIdxBelow = i
			currentIdxAbove = i
			log.Println("Cursor message at thread id ", i)
			break
		}
	}

	lowerBound := cursorY + cursorHeight
	for currentIdxBelow--; currentIdxBelow >= 0 && lowerBound < maxY; currentIdxBelow-- {
		err, msgHeight := m.drawView(0, lowerBound, maxX-1, down, false, thread[currentIdxBelow].UUID, ui) //draw the cursor message
		if err != nil {
			log.Println("error drawing view: ", err)
			return err
		}
		lowerBound += msgHeight
	}
	upperBound := cursorY - 1
	for currentIdxAbove++; currentIdxAbove < len(thread) && upperBound >= 0; currentIdxAbove++ {
		err, msgHeight := m.drawView(0, upperBound, maxX-1, up, false, thread[currentIdxAbove].UUID, ui) //draw the cursor message
		if err != nil {
			log.Println("error drawing view: ", err)
			return err
		}
		upperBound -= msgHeight
	}
	return nil
}

type Direction int

const up Direction = 0
const down Direction = 1

func (h *History) drawView(x, y, w int, dir Direction, isCursor bool, id string, ui *gocui.Gui) (error, int) {
	const borderHeight = 2
	const gutterWidth = 4
	msg := h.Tree.Get(id)
	if msg == nil {
		log.Println("accessed nil message with id:", id)
	}
	numSiblings := len(h.Tree.Children(msg.Parent)) - 1
	contents := wrap.WrapString(msg.Content, uint(w-gutterWidth-1))
	height := strings.Count(contents, "\n") + borderHeight

	var upperLeftX, upperLeftY, lowerRightX, lowerRightY int
	if dir == up {
		upperLeftX = x + gutterWidth
		upperLeftY = y - height
		lowerRightX = x + w
		lowerRightY = y
	} else if dir == down {
		upperLeftX = x + gutterWidth
		upperLeftY = y
		lowerRightX = x + w
		lowerRightY = y + height
	}
	log.Printf("message at (%d,%d) -> (%d,%d)\n", upperLeftX, upperLeftY, lowerRightX, lowerRightY)
	if numSiblings > 0 {
		name := id + "sib"
		if v, err := ui.SetView(name, x, upperLeftY, x+gutterWidth, lowerRightY); err != nil {
			if err != gocui.ErrUnknownView {
				log.Println(err)
				return err, 0
			}
			fmt.Fprintf(v, "%d", numSiblings)
			h.ThreadView.Lock()
			h.ThreadView.ViewIDs[name] = struct{}{}
			h.ThreadView.Unlock()
		}
	}

	if v, err := ui.SetView(id, upperLeftX, upperLeftY, lowerRightX, lowerRightY); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println(err)
			return err, 0
		}
		v.Title = id
		v.Wrap = true
		fmt.Fprint(v, contents)
		if isCursor {
			ui.SetCurrentView(id)
		}
		h.ThreadView.Lock()
		h.ThreadView.ViewIDs[id] = struct{}{}
		h.ThreadView.Unlock()

	}
	return nil, height + 1
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
	id := m.CursorID
	m.ThreadView.Unlock()
	msg := m.Get(id)
	if msg == nil {
		log.Println("Error fetching cursor message: %s", m.CursorID)
		return nil
	} else if msg.Parent == "" {
		log.Println("Cannot move cursor up, nil parent for message: %v", msg)
		return nil
	} else if m.Get(msg.Parent) == nil {
		log.Println("Refusing to move cursor onto nonlocal message with id", msg.Parent)
		return nil
	} else {
		m.ThreadView.Lock()
		m.CursorID = msg.Parent
		m.ThreadView.Unlock()
		return nil
	}
}

func (m *History) CursorLeft(g *gocui.Gui, v *gocui.View) error {
	m.ThreadView.Lock()
	id := m.CursorID
	m.ThreadView.Unlock()
	msg := m.Get(id)
	if msg == nil {
		log.Println("Error fetching cursor message: %s", m.CursorID)
		return nil
	} else if msg.Parent == "" {
		log.Println("Cannot move cursor up, nil parent for message: %v", msg)
		return nil
	} else if len(m.Children(msg.Parent)) < 2 {
		log.Println("Refusing to move cursor onto nonexistent sibling", msg.Parent)
		return nil
	} else {
		siblings := m.Children(msg.Parent)
		var index int
		for i, siblingId := range siblings {
			if siblingId == id {
				index = i
				break
			}
		}
		index = (index + len(siblings) - 1) % len(siblings)
		newCursor := siblings[index]
		log.Printf("Selecting new cursor (old %s) as %s from %v\n", id, newCursor, siblings)
		m.ThreadView.Lock()
		m.ThreadView.LeafID = m.Tree.Leaf(newCursor)
		m.ThreadView.CursorID = newCursor
		log.Println("Selected leaf :", m.ThreadView.LeafID)
		m.ThreadView.Unlock()
		return nil
	}
}

func (m *History) CursorRight(g *gocui.Gui, v *gocui.View) error {
	m.ThreadView.Lock()
	id := m.CursorID
	m.ThreadView.Unlock()
	msg := m.Get(id)
	if msg == nil {
		log.Println("Error fetching cursor message: %s", m.CursorID)
		return nil
	} else if msg.Parent == "" {
		log.Println("Cannot move cursor up, nil parent for message: %v", msg)
		return nil
	} else if len(m.Children(msg.Parent)) < 2 {
		log.Println("Refusing to move cursor onto nonexistent sibling", msg.Parent)
		return nil
	} else {
		siblings := m.Children(msg.Parent)
		var index int
		for i, siblingId := range siblings {
			if siblingId == id {
				index = i
				break
			}
		}
		index = (index + len(siblings) + 1) % len(siblings)
		newCursor := siblings[index]
		log.Printf("Selecting new cursor (old %s) as %s from %v\n", id, newCursor, siblings)
		m.ThreadView.Lock()
		m.ThreadView.LeafID = m.Tree.Leaf(newCursor)
		m.ThreadView.CursorID = newCursor
		log.Println("Selected leaf :", m.ThreadView.LeafID)
		m.ThreadView.Unlock()
		return nil
	}
}

func (m *History) CursorDown(g *gocui.Gui, v *gocui.View) error {
	m.ThreadView.Lock()
	id := m.CursorID
	thread := m.Thread
	m.ThreadView.Unlock()
	msg := m.Get(id)
	if msg == nil {
		log.Println("Error fetching cursor message: %s", m.CursorID)
		return nil
	}
	var prev int = -1
	for i, message := range thread {
		if message.UUID == id {
			prev = i - 1
			break
		}
	}

	if prev >= 0 {
		m.ThreadView.Lock()
		m.CursorID = thread[prev].UUID
		m.ThreadView.Unlock()
	} else {
		log.Println("No previous message")
	}
	return nil
}

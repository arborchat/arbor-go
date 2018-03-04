package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/jroimartin/gocui"
	messages "github.com/whereswaldon/arbor/messages"
)

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: " + os.Args[0] + " <host:port>")
		return
	}
	ui, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Println("Unable to launch ui", err)
		return
	}
	defer ui.Close()

	queries := make(chan string)
	layoutManager := NewList(messages.NewStore(), queries)
	ui.Highlight = true
	ui.Cursor = true
	ui.SelFgColor = gocui.ColorGreen
	ui.SetManager(layoutManager)
	ui.SetCurrentView("message-history")

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Println("Unable to connect", err)
		return
	}
	go handleConn(conn, layoutManager, ui)
	go handleRequests(conn, queries)

	if err := ui.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err = ui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Println("error with ui:", err)
	}
}

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
	const length = 10
	items := make([]string, length)
	current := m.Store.Get(m.LeafID)
	if current == nil {
		return items
	}
	parentID := current.Parent
	for i := length - 1; i >= 0; i-- {
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
	if v, err := ui.SetView("message-history", 0, 0, maxX-1, maxY-5); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println(err)
			return err
		}
		v.Title = "History"
		v.Wrap = true
		v.Autoscroll = true
	}
	history, _ := ui.View("message-history")
	items := m.getItems()
	for _, item := range items {
		fmt.Fprintln(history, item)
	}
	log.Println("rendered history")
	if v, err := ui.SetView("message-input", 0, maxY-4, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println(err)
			return err
		}
		v.Title = "Compose"
		v.Editable = true
		v.Wrap = true
	}
	log.Println("rendered copose")
	return nil
}

func handleConn(conn io.ReadWriteCloser, mlv *MessageListView, ui *gocui.Gui) {
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		log.Println("read ", n, " bytes")
		if err != nil {
			log.Println("unable to read message: ", err)
			return
		}
		a := &messages.ArborMessage{}
		err = json.Unmarshal(data[:n], a)
		if err != nil {
			log.Println("unable to decode message: ", err, string(data))
			continue
		}
		switch a.Type {
		case messages.NEW_MESSAGE:
			// add the new message
			mlv.Add(a.Message)
			mlv.UpdateMessage(a.Message.UUID)
			ui.Update(func(*gocui.Gui) error { return nil })
		default:
			log.Println("Unknown message type: ", string(data))
			continue
		}
	}
}

func handleRequests(conn io.ReadWriteCloser, requestedIds chan string) {
	for id := range requestedIds {
		a := &messages.ArborMessage{
			Type: messages.QUERY,
			Message: &messages.Message{
				UUID: id,
			},
		}
		data, err := json.Marshal(a)
		if err != nil {
			log.Println("Failed to marshal request", err)
			continue
		}
		_, err = conn.Write(data)
		if err != nil {
			log.Println("Failed to write request", err)
			continue
		}
	}
}

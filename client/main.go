package main

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"os"

	ui "github.com/gizak/termui"
	messages "github.com/whereswaldon/arbor/messages"
)

const inputHeight = 5

func main() {
	if len(os.Args) < 2 {
		log.Println("Usage: " + os.Args[0] + " <host:port>")
		return
	}
	err := ui.Init()
	if err != nil {
		log.Println("Unable to launch ui", err)
		return
	}
	defer ui.Close()
	msgList := NewList(messages.NewStore())
	msgList.Overflow = "wrap"
	msgList.Items = []string{
		"test1",
		"test2",
		"test3",
	}
	msgList.BorderLabel = "Messages"
	msgList.Width = ui.TermWidth()
	msgList.Height = ui.TermHeight() - inputHeight
	msgList.Align()

	ui.Render(msgList)

	ui.Handle("/sys/kbd/q", func(ui.Event) {
		ui.StopLoop()
	})

	ui.Handle("/arbor/new_message", func(e ui.Event) {
		log.Println("Got new message")
		msgList.UpdateMessage(e.Data.(string))
		msgList.Align()
		ui.Render(msgList)
	})

	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Println("Unable to connect", err)
		return
	}
	//	io.Copy(os.Stdout, conn)
	log.Println("Opened connection")

	go handleConn(conn, msgList)

	ui.Loop()
}

type MessageListView struct {
	*ui.List
	*messages.Store
}

func NewList(store *messages.Store) *MessageListView {
	return &MessageListView{ui.NewList(), store}
}

func (m *MessageListView) UpdateMessage(id string) {
	msg := m.Store.Get(id)
	m.List.Items = append(m.List.Items, msg.Content)
}

func handleConn(conn io.ReadWriteCloser, mlv *MessageListView) {
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
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
			ui.SendCustomEvt("/arbor/new_message", a.Message.UUID)
		default:
			log.Println("Unknown message type: ", string(data))
			continue
		}
	}
}

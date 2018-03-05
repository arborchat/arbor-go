package clientio

import (
	"encoding/json"
	"io"
	"log"

	"github.com/jroimartin/gocui"
	messages "github.com/whereswaldon/arbor/messages"
)

type MessageView interface {
	Add(*messages.Message)
	UpdateMessage(string)
}

// HandleConn reads from the provided connection and dispatches events to the
// provided MessageListView and Gui as needed to handle new messages coming
// from the server.
func HandleConn(conn io.ReadWriteCloser, mlv MessageView, ui *gocui.Gui) {
	data := make([]byte, 1024)
	for {
		n, err := conn.Read(data)
		log.Println("read ", n, " bytes")
		if err != nil {
			if err == io.EOF {
				log.Println("Connection to server closed, reader shutting down", err)
				break
			}
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

// HandleRequests reads from the requestedIds channel and asks the server to
// send the message details for the messages corresponding to each UUID that
// it receives over the channel.
func HandleRequests(conn io.ReadWriteCloser, requestedIds <-chan string) {
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
			if err == io.EOF {
				log.Println("Connection to server closed, writer shutting down", err)
				break
			}
			log.Println("Failed to write request", err)
			continue
		}
	}
}

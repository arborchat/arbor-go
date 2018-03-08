package clientio

import (
	"encoding/json"
	"io"
	"log"

	messages "github.com/whereswaldon/arbor/messages"
)

// HandleConn reads from the provided connection and writes new messages to the msgs
// channel as they come in.
func HandleNewMessages(conn io.ReadWriteCloser, msgs chan<- *messages.Message) {
	data := make([]byte, 1024)
	defer close(msgs)
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
			msgs <- a.Message
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

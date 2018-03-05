package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"

	messages "github.com/whereswaldon/arbor/messages"
)

const replyThreshold = 0.5

func main() {
	if len(os.Args) < 2 {
		log.Fatalln("Usage: " + os.Args[0] + " <host:port>")
	}
	conn, err := net.Dial("tcp", os.Args[1])
	if err != nil {
		log.Fatalln("Unable to connect", err)
		return
	}

	data := make([]byte, 1024)
	replyCounter := 0
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
			// choose whether to reply
			if a.Message.UUID != "" && rand.Float64() < replyThreshold {
				log.Println("Choosing to reply to ", a.Message.UUID)
				a := &messages.ArborMessage{
					Type: messages.NEW_MESSAGE,
					Message: &messages.Message{
						Parent:  a.Message.Parent,
						Content: fmt.Sprintf("%d", replyCounter),
					},
				}
				replyCounter++
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
		default:
			log.Println("Unknown message type: ", string(data))
			continue
		}
	}
}

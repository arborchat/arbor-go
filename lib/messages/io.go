package messages

import (
	"encoding/json"
	"io"
	"log"
)

func MakeMessageWriter(conn io.ReadWriteCloser) chan<- *ArborMessage {
	input := make(chan *ArborMessage)
	go func() {
		for message := range input {
			data, err := json.Marshal(message)
			if err != nil {
				log.Println("Error marshalling welcome", err)
				return
			}
			_, err = conn.Write(data)
			if err != nil {
				log.Println("Error sending welcome", err)
				return
			}
		}
	}()
	return input
}

func MakeMessageReader(conn io.ReadWriteCloser) <-chan *ArborMessage {
	output := make(chan *ArborMessage)
	data := make([]byte, 1024)
	go func() {
		defer close(output)
		for {
			n, err := conn.Read(data)
			if err != nil {
				log.Println("Error reading from conn:", err)
				return
			}
			a := &ArborMessage{}
			err = json.Unmarshal(data[:n], a)
			if err != nil {
				log.Println("Error decoding json:", err)
				continue
			}
			output <- a
		}
	}()
	return output
}

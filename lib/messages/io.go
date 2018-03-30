package messages

import (
	"encoding/json"
	"io"
	"log"
)

func MakeMessageWriter(conn io.ReadWriteCloser) chan<- *ArborMessage {
	input := make(chan *ArborMessage)
	go func() {
		encoder := json.NewEncoder(conn)
		for message := range input {
			err := encoder.Encode(message)
			if err != nil {
				log.Println("Error encoding message", err)
				return
			}
		}
	}()
	return input
}

func MakeMessageReader(conn io.ReadWriteCloser) <-chan *ArborMessage {
	output := make(chan *ArborMessage)
	go func() {
		defer close(output)
		decoder := json.NewDecoder(conn)
		for {
			a := &ArborMessage{}
			err := decoder.Decode(a)
			if err != nil {
				log.Println("Error decoding json:", err)
				return
			}
			output <- a
		}
	}()
	return output
}

package arbor

import (
	"encoding/json"
	"io"
	"log"
)

// MakeMessageWriter wraps the io.ReadCloser and returns a channel of
// ProtocolMessage pointers. Any ArborMessage sent over that channel will be
// written onto the io.ReadCloser as JSON. This function handles all
// marshalling.
func MakeMessageWriter(conn io.WriteCloser) chan<- *ProtocolMessage {
	input := make(chan *ProtocolMessage)
	go func() {
		defer close(input)
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

// MakeMessageReader wraps the io.ReadCloser and returns a channel of
// ProtocolMessage pointers. Any JSON received over the io.ReadCloser will
// be unmarshalled into an ProtocolMessage struct and sent over the returned
// channel.
func MakeMessageReader(conn io.ReadCloser) <-chan *ProtocolMessage {
	output := make(chan *ProtocolMessage)
	go func() {
		defer close(output)
		decoder := json.NewDecoder(conn)
		for {
			a := &ProtocolMessage{}
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

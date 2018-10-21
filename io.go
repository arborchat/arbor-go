package arbor

import (
	"encoding/json"
	"io"
	"log"
)

// MakeMessageWriter wraps the io.Writer and returns a channel of
// ProtocolMessage pointers. Any ProtocolMessage sent over that channel will be
// written onto the io.Writer as JSON. This function handles all
// marshalling.
func MakeMessageWriter(conn io.Writer) chan<- *ProtocolMessage {
	input := make(chan *ProtocolMessage)
	go func() {
		defer close(input)
		encoder := json.NewEncoder(conn)
		for message := range input {
			err := encoder.Encode(message)
			if err != nil {
				if err == io.EOF {
					log.Println("Writer connection closed", err)
					return
				}
				log.Println("Error encoding message", err)
			}
		}
	}()
	return input
}

// MakeMessageReader wraps the io.Reader and returns a channel of
// ProtocolMessage pointers. Any JSON received over the io.Reader will
// be unmarshalled into an ProtocolMessage struct and sent over the returned
// channel.
func MakeMessageReader(conn io.Reader) <-chan *ProtocolMessage {
	output := make(chan *ProtocolMessage)
	go func() {
		defer close(output)
		decoder := json.NewDecoder(conn)
		for {
			a := &ProtocolMessage{}
			err := decoder.Decode(a)
			if err != nil {
				if err == io.EOF {
					log.Println("Reader connection closed", err)
					return
				}
				log.Println("Error decoding json:", err)
			}
			output <- a
		}
	}()
	return output
}

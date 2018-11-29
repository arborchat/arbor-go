package arbor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
)

// Reader defines the behavior of types that can emit arbor protocol messages
type Reader interface {
	// Populates the provided ProtocolMessage pointer with the contents of a newly read message
	Read(*ProtocolMessage) error
}

// Writer defines the behavior of types that can consume arbor protocol messages
type Writer interface {
	// Consumes the provided ProtocolMessage without modifying it
	Write(*ProtocolMessage) error
}

// ReadWriter defines the behavior of types that can both emit and consume arbor
// protocol messages
type ReadWriter interface {
	Reader
	Writer
}

// ReadWriteCloser defines the behavior of types that can both emit and consume arbor
// protocol messages that have a logical "Close" operation (file/socket wrappers, for instance)
type ReadWriteCloser interface {
	ReadWriter
	io.Closer
}

// ProtocolReader reads arbor protocol messages (as JSON) from an io.Reader
type ProtocolReader struct {
	in  chan *ProtocolMessage
	out chan error
}

// ensure ProtocolReader always fulfills the Reader interface
var _ Reader = &ProtocolReader{}

// NewProtocolReader wraps the source to make serializing *ProtocolMessages easy.
func NewProtocolReader(source io.Reader) (*ProtocolReader, error) {
	if source == nil {
		return nil, fmt.Errorf("NewProtocolReader cannot wrap nil")
	}
	if reflect.ValueOf(source).IsNil() {
		return nil, fmt.Errorf("NewProtocolReader given io.Reader typed nil")
	}
	reader := &ProtocolReader{
		in:  make(chan *ProtocolMessage),
		out: make(chan error),
	}
	go reader.readLoop(source)
	return reader, nil
}

func (r *ProtocolReader) readLoop(conn io.Reader) {
	defer close(r.out)
	decoder := json.NewDecoder(conn)
	for msg := range r.in {
		r.out <- decoder.Decode(msg)
	}
}

// Read attempts to read a JSON-serialized ProtocolMessage from the Reader's source
// into the provided ProtocolMessage. If the provided message is nil, it will error.
// This method will block until a ProtocolMessage becomes available.
func (r *ProtocolReader) Read(into *ProtocolMessage) error {
	r.in <- into
	return <-r.out
}

// ProtocolWriter writes arbor protocol messages (as JSON) to an io.Reader
type ProtocolWriter struct{}

// ensure that ProtocolWriter satisfies the Writer interface at compile-time
var _ Writer = &ProtocolWriter{}

// NewProtocolWriter creates a ProtocolWriter by wrapping a destination io.Writer
func NewProtocolWriter(destination io.Writer) (*ProtocolWriter, error) {
	return nil, nil
}

// Write persists the given arbor protocol message into the ProtocolWriter's backing
// io.Writer
func (w *ProtocolWriter) Write(from *ProtocolMessage) error {
	return nil
}

// ProtocolReadWriter can read and write arbor protocol messages (as JSON) from an io.ReadWriter
type ProtocolReadWriter struct {
	ProtocolReader
	ProtocolWriter
}

// ProtocolReadWriteCloser can read and write arbor protocol messages (as JSON) from an io.ReadWriteCloser
type ProtocolReadWriteCloser struct {
	ProtocolReadWriter
}

// MakeMessageWriter wraps the io.Writer and returns a channel of
// ProtocolMessage pointers. Any ProtocolMessage sent over that channel will be
// written onto the io.Writer as JSON. This function handles all
// marshalling. If a message fails to marshal for any reason, or if a write error
// occurs, the returned channel will be closed and no further messages will be
// written to the io.Writer.
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
				return
			}
		}
	}()
	return input
}

// MakeMessageReader wraps the io.ReadCloser and returns a channel of
// ProtocolMessage pointers. Any JSON received over the io.ReadCloser will
// be unmarshalled into an ProtocolMessage struct and sent over the returned
// channel. If invalid JSON is received, the ReadCloser will close the io.ReadCloser
// and the returned channel.
func MakeMessageReader(conn io.ReadCloser) <-chan *ProtocolMessage {
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
				// if we received unparsable JSON, just hang up.
				defer func() {
					if closeErr := conn.Close(); closeErr != nil {
						log.Println("Error closing connection:", closeErr)
					}
				}()

				log.Println("Error decoding json, hanging up:", err)
				return
			}
			output <- a
		}
	}()
	return output
}

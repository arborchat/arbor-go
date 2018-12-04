package arbor

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"sync"
)

type closeable struct {
	io.ReadWriter
}

func (c closeable) Close() error {
	return nil
}

// NoopRWCloser wraps an io.ReadWriter with an implementation of io.Closer's
// Close() method that does nothing. It's like ioutil.NoopCloser but for
// io.ReadWriters instead of just io.Readers.
func NoopRWCloser(in io.ReadWriter) io.ReadWriteCloser {
	return closeable{in}
}

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
	closed bool
	sync.RWMutex
	in  chan *ProtocolMessage
	out chan error
}

// ensure ProtocolReader always fulfills the Reader interface
var _ Reader = &ProtocolReader{}

func isNilPointer(in interface{}) bool {
	return reflect.ValueOf(in).Kind() == reflect.Ptr && reflect.ValueOf(in).IsNil()
}

// NewProtocolReader wraps the source to make serializing *ProtocolMessages easy.
func NewProtocolReader(source io.Reader) (*ProtocolReader, error) {
	if source == nil {
		return nil, fmt.Errorf("NewProtocolReader cannot wrap nil")
	}
	if isNilPointer(source) {
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
	r.RLock()
	defer r.RUnlock()
	if r.closed {
		return fmt.Errorf("Reading from closed reader")
	}
	r.in <- into
	return <-r.out
}

func (r *ProtocolReader) stop() {
	r.Lock()
	defer r.Unlock()
	close(r.in)
}

// ProtocolWriter writes arbor protocol messages (as JSON) to an io.Reader
type ProtocolWriter struct {
	sync.RWMutex
	closed    bool
	toWrite   chan *ProtocolMessage
	writeErrs chan error
}

// ensure that ProtocolWriter satisfies the Writer interface at compile-time
var _ Writer = &ProtocolWriter{}

// NewProtocolWriter creates a ProtocolWriter by wrapping a destination io.Writer
func NewProtocolWriter(destination io.Writer) (*ProtocolWriter, error) {
	if destination == nil {
		return nil, fmt.Errorf("NewProtocolWriter cannot wrap nil")
	}
	if isNilPointer(destination) {
		return nil, fmt.Errorf("NewProtocolWriter given io.Writer typed nil")
	}
	writer := &ProtocolWriter{
		toWrite:   make(chan *ProtocolMessage),
		writeErrs: make(chan error),
	}
	go writer.writeLoop(destination)
	return writer, nil
}

func (w *ProtocolWriter) writeLoop(conn io.Writer) {
	defer close(w.writeErrs)
	encoder := json.NewEncoder(conn)
	for msg := range w.toWrite {
		w.writeErrs <- encoder.Encode(msg)
	}
}

func (w *ProtocolWriter) stop() {
	w.Lock()
	defer w.Unlock()
	close(w.toWrite)
}

// Write persists the given arbor protocol message into the ProtocolWriter's backing
// io.Writer
func (w *ProtocolWriter) Write(target *ProtocolMessage) error {
	if target == nil {
		return fmt.Errorf("Cannot write nil message")
	}
	w.RLock()
	defer w.RUnlock()
	if w.closed {
		return fmt.Errorf("Cannot write into closed Writer")
	}
	w.toWrite <- target
	return <-w.writeErrs
}

// ProtocolReadWriter can read and write arbor protocol messages (as JSON) from an io.ReadWriter
type ProtocolReadWriter struct {
	*ProtocolReader
	*ProtocolWriter
	closeReq chan struct{}
	closeRes chan error
}

// Ensure that ProtocolReadWriteCloser statisfies ReadWriteCloser at compile time
var _ ReadWriteCloser = &ProtocolReadWriter{}

// NewProtocolReadWriter wraps the given io.ReadWriter so that it is possible to both read
// and write arbor protocol messages to it.
func NewProtocolReadWriter(wrap io.ReadWriteCloser) (*ProtocolReadWriter, error) {
	reader, err := NewProtocolReader(wrap)
	if err != nil {
		return nil, err
	}
	writer, err := NewProtocolWriter(wrap)
	if err != nil {
		return nil, err
	}
	rw := &ProtocolReadWriter{
		ProtocolReader: reader,
		ProtocolWriter: writer,
		closeReq:       make(chan struct{}),
		closeRes:       make(chan error),
	}
	go rw.closeWait(wrap)
	return rw, nil
}

func (c *ProtocolReadWriter) closeWait(target io.Closer) {
	defer close(c.closeRes)
	<-c.closeReq
	c.ProtocolReader.stop()
	c.ProtocolWriter.stop()
	c.closeRes <- target.Close()
}

// Close both closes the io.ReadWriteCloser wrapped by this ProtocolReadWriter and tears down all
// protocol-related internal structure. Once you close a ProtocolReadWriter, you must create a new
// one in order to use it again.
func (c *ProtocolReadWriter) Close() (err error) {
	defer func() {
		recovered := recover()
		if recovered == nil {
			err = nil
		} else {
			err = recovered.(error)
		}
	}()
	close(c.closeReq) // signal that we should shut everything down
	return <-c.closeRes
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

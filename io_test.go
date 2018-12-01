package arbor_test

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/jordwest/mock-conn"
)

const (
	welcomeExample = "{\"Type\":0,\"Root\":\"f4ae0b74-4025-4810-41d6-5148a513c580\",\"Recent\":[\"92d24e9d-12cc-4742-6aaf-ea781a6b09ec\",\"880be029-0d7c-4a3f-558d-d90bf79cbc1d\"],\"Major\":0,\"Minor\":1}"
	newExample     = "{\"Type\":2,\"UUID\":\"92d24e9d-12cc-4742-6aaf-ea781a6b09ec\",\"Parent\":\"f4ae0b74-4025-4810-41d6-5148a513c580\",\"Content\":\"A riveting example message.\",\"Username\":\"Examplius_Caesar\",\"Timestamp\":1537738224}"
	queryExample   = "{\"Type\":1,\"UUID\":\"f4ae0b74-4025-4810-41d6-5148a513c580\"}"
)

// TestNilReader ensures that NewProtocolReader correctly handles being provided with a
// nil io.Reader
func TestNilReader(t *testing.T) {
	reader, err := arbor.NewProtocolReader(nil)
	if err == nil {
		t.Error("NewProtocolReader should error when given a nil io.Reader")
	}
	if reader != nil {
		t.Error("NewProtocolReader should return nil ProtocolReader when given a nil io.Reader")
	}
}

type badReadWriteCloser struct{ Field int }

func (b *badReadWriteCloser) Read([]byte) (int, error) {
	return b.Field, nil // access a property to trigger a nil pointer dereference
}

func (b *badReadWriteCloser) Write([]byte) (int, error) {
	return b.Field, nil // access a property to trigger a nil pointer dereference
}

func (b *badReadWriteCloser) Close() error {
	return nil
}

type detectableCloser struct {
	io.ReadWriter
	closed bool
}

var _ io.ReadWriteCloser = &detectableCloser{}

func (c *detectableCloser) Close() error {
	c.closed = true
	return nil
}

// TestTypedNilReader ensures that NewProtocolReader correctly handles being provided with a
// nil concrete value with a non-nil concrete type wrapped in the io.Reader interface.
func TestTypedNilReader(t *testing.T) {
	// create a typed nil
	var bad *badReadWriteCloser
	var typedBad io.Reader = bad
	reader, err := arbor.NewProtocolReader(typedBad)
	if err == nil {
		t.Error("NewProtocolReader should error when given a nil io.Reader")
	}
	if reader != nil {
		t.Error("NewProtocolReader should return nil ProtocolReader when given a nil io.Reader")
	}
}

// TestReaderRead ensures that we can read a message out of a ProtocolReader when
// it is given proper input.
func TestReaderRead(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	welcome := getWelcome()
	err := encoder.Encode(welcome)
	if err != nil {
		t.Skip("Unable to write test data", err)
	}
	reader, err := arbor.NewProtocolReader(buf)
	if err != nil {
		t.Error("Unable to construct Reader with valid input", err)
	} else if reader == nil {
		t.Error("Got nil Reader back when invoking constructor with valid input")
	}
	proto := arbor.ProtocolMessage{}
	err = reader.Read(&proto)
	if err != nil {
		t.Error("Expected to be able to read message from buffer", err)
	}
	if proto.Type != welcome.Type || proto.Root != welcome.Root || proto.Major != welcome.Major || proto.Minor != welcome.Minor {
		t.Errorf("Expected %v, found %v", welcome, proto)
	}
	for i := 0; i < len(welcome.Recent) && i < len(proto.Recent); i++ {
		if welcome.Recent[i] != proto.Recent[i] {
			t.Errorf("Recents don't match, expected %v found %v", welcome.Recent, proto.Recent)
		}
	}
}

// TestReaderReadNil ensures that we properly handle nil input to Read.
func TestReaderReadNil(t *testing.T) {
	buf := new(bytes.Buffer)
	encoder := json.NewEncoder(buf)
	welcome := getWelcome()
	err := encoder.Encode(welcome)
	if err != nil {
		t.Skip("Unable to write test data", err)
	}
	reader, err := arbor.NewProtocolReader(buf)
	if err != nil {
		t.Error("Unable to construct Reader with valid input", err)
	} else if reader == nil {
		t.Error("Got nil Reader back when invoking constructor with valid input")
	}
	err = reader.Read(nil)
	if err == nil {
		t.Error("Expected an error from trying to read into nil pointer")
	}
}

// TestNilWriter ensures that NewProtocolWriter correctly handles being provided with a
// nil io.Writer
func TestNilWriter(t *testing.T) {
	reader, err := arbor.NewProtocolWriter(nil)
	if err == nil {
		t.Error("NewProtocolWriter should error when given a nil io.Writer")
	}
	if reader != nil {
		t.Error("NewProtocolWriter should return nil ProtocolWriter when given a nil io.Writer")
	}
}

// TestTypedNilWriter ensures that NewProtocolWriter correctly handles being provided with a
// nil concrete value with a non-nil concrete type wrapped in the io.Writer interface.
func TestTypedNilWriter(t *testing.T) {
	// create a typed nil
	var bad *badReadWriteCloser
	var typedBad io.Writer = bad
	reader, err := arbor.NewProtocolWriter(typedBad)
	if err == nil {
		t.Error("NewProtocolWriter should error when given a nil io.Writer")
	}
	if reader != nil {
		t.Error("NewProtocolWriter should return nil ProtocolWriter when given a nil io.Writer")
	}
}

// TestWriterWrite ensures that we can write a message with a ProtocolWriter when
// it is given proper input.
func TestWriterWrite(t *testing.T) {
	buf := new(bytes.Buffer)
	writer, err := arbor.NewProtocolWriter(buf)
	if err != nil {
		t.Error("Unable to construct Writer with valid input", err)
	} else if writer == nil {
		t.Error("Got nil Writer back when invoking constructor with valid input")
	}
	welcome := getWelcome()
	err = writer.Write(welcome)
	if err != nil {
		t.Error("Unable to write message into valid destination", err)
	}
	proto := arbor.ProtocolMessage{}
	decoder := json.NewDecoder(buf)
	err = decoder.Decode(&proto)
	if err != nil {
		t.Skip("Unable to read test data", err)
	}
	if proto.Type != welcome.Type || proto.Root != welcome.Root || proto.Major != welcome.Major || proto.Minor != welcome.Minor {
		t.Errorf("Expected %v, found %v", welcome, proto)
	}
	for i := 0; i < len(welcome.Recent) && i < len(proto.Recent); i++ {
		if welcome.Recent[i] != proto.Recent[i] {
			t.Errorf("Recents don't match, expected %v found %v", welcome.Recent, proto.Recent)
		}
	}
}

// TestWriterWriteNil ensures that we properly handle nil input to Write.
func TestWriterWriteNil(t *testing.T) {
	buf := new(bytes.Buffer)
	writer, err := arbor.NewProtocolWriter(buf)
	if err != nil {
		t.Error("Unable to construct Writer with valid input", err)
	} else if writer == nil {
		t.Error("Got nil Writer back when invoking constructor with valid input")
	}
	err = writer.Write(nil)
	if err == nil {
		t.Error("Expected an error from trying to read into nil pointer")
	}
}

// TestReadWriter ensures that NewProtocolReadWriter returns a valid ProtocolReadWriter when given
// valid input.
func TestReadWriter(t *testing.T) {
	buf := new(bytes.Buffer)
	rw, err := arbor.NewProtocolReadWriter(arbor.NoopRWCloser(buf))
	if err != nil {
		t.Error("NewProtocolReadWriter should not error when given a valid io.ReadWriter")
	}
	if rw == nil {
		t.Error("NewProtocolReadWriter should not return nil when given a valid io.ReadWriter")
	}
}

// TestNilReadWriter ensures that NewProtocolReadWriter correctly handles being provided with a
// nil io.ReadWriter
func TestNilReadWriter(t *testing.T) {
	rw, err := arbor.NewProtocolReadWriter(nil)
	if err == nil {
		t.Error("NewProtocolReadWriter should error when given a nil io.ReadWriter")
	}
	if rw != nil {
		t.Error("NewProtocolReadWriter should return nil ProtocolReadWriter when given a nil io.ReadWriter")
	}
}

// TestTypedNilReadWriter ensures that NewProtocolReadWriter correctly handles being provided with a
// nil concrete value with a non-nil concrete type wrapped in the io.ReadWriter interface.
func TestTypedNilReadWriter(t *testing.T) {
	// create a typed nil
	var bad *badReadWriteCloser
	var typedBad io.ReadWriteCloser = bad
	rw, err := arbor.NewProtocolReadWriter(typedBad)
	if err == nil {
		t.Error("NewProtocolReadWriter should error when given a nil io.ReadWriter")
	}
	if rw != nil {
		t.Error("NewProtocolReadWriter should return nil ProtocolReadWriter when given a nil io.ReadWriter")
	}
}

// TestProtocolRWClose ensures that the Close method on ProtocolReadWriteClosers closes the underlying
// io.ReadWriteCloser and returns an error when you attempt to close it more than once.
func TestProtocolRWClose(t *testing.T) {
	detect := &detectableCloser{ReadWriter: new(bytes.Buffer)}
	closer, err := arbor.NewProtocolReadWriter(detect)
	if err != nil {
		t.Skip("Failed to construct ProtocolReadWriteCloser", err)
	}
	if detect.closed {
		t.Error("ProtocolReadWriteCloser should not close underlying io.ReadWriteCloser unasked")
	}
	err = closer.Close()
	if err != nil {
		t.Error("Should have been able to close open underlying io.ReadWriteCloser", err)
	}
	if !detect.closed {
		t.Error("After Close(), underlying io.ReadWriteCloser should be closed")
	}
	err = closer.Close()
	if err == nil {
		t.Error("closing underlying io.ReadWriteCloser twice should have errored")
	}

}

// TestMakeMessageReader checks that MakeMessageReader properly reads messages
// from the input connection.
func TestMakeMessageReader(t *testing.T) {
	testMsgs := []string{
		newExample + "\n",
		welcomeExample + "\n",
		queryExample + "\n",
	}
	conn := mock_conn.NewConn()
	recvChan := arbor.MakeMessageReader(conn.Client)
	server := conn.Server
	for _, msg := range testMsgs {
		testMsg := []byte(msg)
		n, err := server.Write(testMsg)
		if err != nil || n != len(testMsg) {
			t.Skipf("Unable to write message \"%s\" into mock connection", msg)
		}
		parsed := <-recvChan
		if parsed == nil {
			t.Error("MakeMessageReader sent nil ProtocolMessage")
		}
	}
}

// TestMakeMessageReaderInvalid checks that MakeMessageReader hangs up when it recieves bad
// input.
func TestMakeMessageReaderInvalid(t *testing.T) {
	testMsgs := []string{
		string([]byte{0x1b}) + "\n", // this is the ASCII escape character
	}

	for _, msg := range testMsgs {
		conn := mock_conn.NewConn()
		recvChan := arbor.MakeMessageReader(conn.Client)
		server := conn.Server
		testMsg := []byte(msg)
		n, err := server.Write(testMsg)
		if err != nil || n != len(testMsg) {
			t.Skipf("Unable to write message \"%s\" into mock connection", msg)
		}
		parsed := <-recvChan
		if parsed != nil {
			t.Error("MakeMessageReader did not close output channel on bad input")
		}
		if n, err = server.Write(testMsg); err == nil {
			// on a tcp connection, we'd get io.EOF, but our mock doesn't work that way.
			// just need to make sure that we get an error
			t.Error("MakeMessageReader failed to close the connection on bad input, no error on write")
		} else if n > 0 {
			t.Error("MakeMessageReader failed to close connection on bad input, able to write data")
		}
	}
}

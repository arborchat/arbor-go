package arbor_test

import (
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

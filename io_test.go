package arbor_test

import (
	"testing"

	arbor "github.com/arborchat/arbor-go"
	"github.com/jordwest/mock-conn"
)

// TestMakeMessageReader checks that MakeMessageReader properly reads messages
// from the input io.Reader.
func TestMakeMessageReader(t *testing.T) {
	testMsgs := []string{
		"{}\n",
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

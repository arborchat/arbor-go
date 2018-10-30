package arbor_test

import (
	"testing"
	"time"

	arbor "github.com/arborchat/arbor-go"
)

const testContent = "Test message"
const testContent2 = "another test message"

// TestNewChatMessage ensures that the chat message is created with its
// `Content` and `Timestamp` fields populated and its `Parent` field cleared.
func TestNewChatMessage(t *testing.T) {
	m, err := arbor.NewChatMessage(testContent)
	if err != nil {
		t.Fatalf("Unable to create ChatMessage with valid parameters, error: %v", err)
	}
	if m == nil {
		t.Fatal("Recieved nil ChatMessage from valid call to NewChatMessage")
	}
	if m.Parent != "" {
		t.Errorf("Expected \"\" as Parent, found %v", m.Parent)
	}
	if m.Content != testContent {
		t.Errorf("Expected \"%s\" as Content, found \"%s\"", testContent, m.Content)
	}
	if unreasonableTimestamp(m.Timestamp) {
		t.Errorf("Timestamp invalid (either in the future or before the epoch): %d", m.Timestamp)
	}
}

func newMessageOrSkip(t *testing.T, content string) *arbor.ChatMessage {
	m, err := arbor.NewChatMessage(testContent)
	if err != nil || m == nil {
		t.Skip("Unable to create message")
	}
	return m
}

// TestAssignID ensures that the AssignID function actually populates the UUID field
// of the ChatMessage
func TestAssignID(t *testing.T) {
	m := newMessageOrSkip(t, testContent)
	err := m.AssignID()
	if err != nil {
		t.Error("Failed to assign UUID", err)
		return
	}
	if m.UUID == "" {
		t.Error("No UUID assigned", m)
	}
}

func unreasonableTimestamp(timestamp int64) bool {
	return timestamp > time.Now().Unix() || timestamp < 0
}

// TestReply ensures that the Reply method creates a new ChatMessage with the correct
// Parent and Content as well as a reasonable Timestamp.
func TestReply(t *testing.T) {
	m := newMessageOrSkip(t, testContent)
	err := m.AssignID()
	if err != nil || m.UUID == "" {
		t.Skip("Failed to assign UUID", err)
		return
	}
	m2, err := m.Reply(testContent2)
	if err != nil {
		t.Error("Failed create reply", err)
		return
	}
	if m2.Parent == "" {
		t.Error("No parent assigned", m)
	} else if m2.Parent != m.UUID {
		t.Errorf("Expected Parent to be %s, found %s", m.UUID, m2.Parent)
	}
	if m2.Content != testContent2 {
		t.Errorf("Expected Content to be \"%s\", found \"%s\"", testContent2, m2.Content)
	}
	if m == m2 {
		t.Error("Reply reused original structure")
	}
	if unreasonableTimestamp(m2.Timestamp) {
		t.Errorf("Timestamp invalid (either in the future or before the epoch): %d", m2.Timestamp)
	}
}

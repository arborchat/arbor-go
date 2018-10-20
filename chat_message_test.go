package arbor_test

import (
	"testing"
	"time"

	arbor "github.com/arborchat/arbor-go"
)

const testContent = "Test mesage"

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
	if m.Timestamp > time.Now().Unix() || m.Timestamp < 0 {
		t.Errorf("Timestamp invalid (either in the future or before the epoch): %d", m.Timestamp)
	}
}

// TestAssignID ensures that the AssignID function actually populates the UUID field
// of the ChatMessage
func TestAssignID(t *testing.T) {
	m, err := arbor.NewChatMessage(testContent)
	if err != nil || m == nil {
		t.Skip("Unable to create message")
	}
	err = m.AssignID()
	if err != nil {
		t.Error("Failed to assign UUID", err)
		return
	}
	if m.UUID == "" {
		t.Error("No UUID assigned", m)
	}
}

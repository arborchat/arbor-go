package arbor_test

import (
	"math/rand"
	"strconv"
	"testing"

	arbor "github.com/arborchat/arbor-go"
)

// TestNewStore ensures that NewStore returns a store.
func TestNewStore(t *testing.T) {
	s := arbor.NewStore()
	if s == nil {
		t.Error("NewStore() returned a nil store")
	}
}

func randomMessage() *arbor.ChatMessage {
	m, _ := arbor.NewChatMessage(strconv.Itoa(rand.Int()))
	_ = m.AssignID()
	m.Username = strconv.Itoa(rand.Int())
	return m
}

const iterations = 10000

// TestAddAndGet ensures that any message inserted into the store with "Add" can be
// retrieved with Get.
func TestAddAndGet(t *testing.T) {
	s := arbor.NewStore()
	if s == nil {
		t.Skip("Got nil store")
	}
	msgs := make([]*arbor.ChatMessage, iterations)
	for i := 0; i < iterations; i++ {
		msgs[i] = randomMessage()
		s.Add(msgs[i])
	}
	for i := 0; i < iterations; i++ {
		m := s.Get(msgs[i].UUID)
		if m == nil {
			t.Error("Failed to retrieve message previously stored")
		}
		if m.UUID != msgs[i].UUID {
			t.Errorf("Expected %s, found %s", msgs[i].UUID, m.UUID)
		}
		if m.Parent != msgs[i].Parent {
			t.Errorf("Expected %s, found %s", msgs[i].Parent, m.Parent)
		}
		if m.Timestamp != msgs[i].Timestamp {
			t.Errorf("Expected %d, found %d", msgs[i].Timestamp, m.Timestamp)
		}
		if m.Username != msgs[i].Username {
			t.Errorf("Expected %s, found %s", msgs[i].Username, m.Username)
		}
		if m.Content != msgs[i].Content {
			t.Errorf("Expected %s, found %s", msgs[i].Content, m.Content)
		}
	}
}

const nonexsitentID = "nonexistent"

// TestGetNotAdded ensures that Get() returns nil when you ask for a message that has
// never been Add()-ed
func TestGetNotAdded(t *testing.T) {
	s := arbor.NewStore()
	if s == nil {
		t.Skip("Got nil store")
	}
	m := s.Get(nonexsitentID)
	if m != nil {
		t.Error("Recieved non-nil message when getting a non-existent message ID", m)
	}
}

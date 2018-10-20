package arbor_test

import (
	"strings"
	"testing"
	"time"

	arbor "github.com/arborchat/arbor-go"
)

const (
	testRoot           = "root"
	testID1            = "1"
	testID2            = "2"
	testID3            = "3"
	testID4            = "4"
	testUser           = "testopheles"
	testBadMessageType = 99
)

func getWelcome() *arbor.ProtocolMessage {
	return &arbor.ProtocolMessage{
		Type:   arbor.WelcomeType,
		Root:   testRoot,
		Recent: []string{testID1, testID2, testID3, testID4},
		Major:  0,
		Minor:  1,
	}
}

func getNew() *arbor.ProtocolMessage {
	return &arbor.ProtocolMessage{
		Type: arbor.NewMessageType,
		ChatMessage: &arbor.ChatMessage{
			UUID:      testID1,
			Parent:    testID2,
			Username:  testUser,
			Timestamp: time.Now().Unix(),
			Content:   testContent,
		},
	}
}
func getQuery() *arbor.ProtocolMessage {
	return &arbor.ProtocolMessage{
		Type: arbor.QueryType,
		ChatMessage: &arbor.ChatMessage{
			UUID: testID1,
		},
	}
}

func getInvalid() *arbor.ProtocolMessage {
	return &arbor.ProtocolMessage{Type: testBadMessageType}
}

// TestMarshalJSON ensures that the JSON serialization for each message type works
// and does not include message fields irrelevant for that message type.
func TestMarshalJSON(t *testing.T) {
	welcomeOnlyFields := []string{"Root", "Recent", "Major", "Minor"}
	welcomeAllFields := append(welcomeOnlyFields, "Type")
	newOnlyFields := []string{"Content", "Parent", "Username", "Timestamp"}
	newAllFields := append(newOnlyFields, "Type", "UUID")
	queryAllFields := []string{"Type", "UUID"}
	welcome := getWelcome()
	byteWelcome, err := welcome.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	strWelcome := string(byteWelcome)
	if strWelcome == "" {
		t.Error("Received empty string")
	}
	if containsAnyOf(strWelcome, newOnlyFields) {
		t.Error("Welcome message contained fields from new message", strWelcome)
	}
	if !containsEachOf(strWelcome, welcomeAllFields) {
		t.Error("Welcome message did not contain all welcome message fields", strWelcome)
	}
	newmessage := getNew()
	byteNewmessage, err := newmessage.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	strNewmessage := string(byteNewmessage)
	if strNewmessage == "" {
		t.Error("Received empty string")
	}
	if containsAnyOf(strNewmessage, welcomeOnlyFields) {
		t.Error("New message contained fields from welcome message", strNewmessage)
	}
	if !containsEachOf(strNewmessage, newAllFields) {
		t.Error("New message did not contain all new message fields", strNewmessage)
	}
	query := getQuery()
	byteQuery, err := query.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	strQuery := string(byteQuery)
	if strQuery == "" {
		t.Error("Received empty string")
	}
	if containsAnyOf(strQuery, welcomeOnlyFields) {
		t.Error("Query message contained fields from welcome message", strQuery)
	}
	if containsAnyOf(strQuery, newOnlyFields) {
		t.Error("Query message contained fields from new message", strQuery)
	}
	if !containsEachOf(strQuery, queryAllFields) {
		t.Error("New message did not contain all new message fields", strQuery)
	}
	badOutput, err := getInvalid().MarshalJSON()
	if err == nil {
		t.Error("Marshalling bad message type should be error")
	}
	if badOutput != nil {
		t.Error("Marshalling bad message type should return nil slice, got", badOutput)
	}
}

// TestString ensures that String() returns an non-empty string for valid messages and
// and empty string for invalid ones
func TestString(t *testing.T) {
	msgFuncs := []func() *arbor.ProtocolMessage{getNew, getQuery, getWelcome}
	for _, function := range msgFuncs {
		m := function()
		s := m.String()
		if s == "" {
			t.Error("String should not return empty string for valid message", m)
		}
	}
	m := getInvalid()
	s := m.String()
	if s != "" {
		t.Error("String should be empty for invalid message", m)
	}
}

func containsAnyOf(input string, targets []string) bool {
	for _, s := range targets {
		if strings.Contains(input, s) {
			return true
		}
	}
	return false
}
func containsEachOf(input string, targets []string) bool {
	count := 0
	for _, s := range targets {
		if strings.Contains(input, s) {
			count++
		}
	}
	return count == len(targets)
}

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

func getMeta() *arbor.ProtocolMessage {
	return &arbor.ProtocolMessage{
		Type: arbor.MetaType,
		Meta: map[string]string{"key": "value"},
	}
}

func getInvalid() *arbor.ProtocolMessage {
	return &arbor.ProtocolMessage{Type: testBadMessageType}
}

func contains(input string, noneOf []string, eachOf []string) bool {
	return !containsAnyOf(input, noneOf) && !containsEachOf(input, eachOf)
}

var (
	// define fields that are only legal in specific message types for use in
	// serialization tests
	welcomeOnlyFields = []string{"Root", "Recent", "Major", "Minor"}
	welcomeAllFields  = append(welcomeOnlyFields, "Type")
	newOnlyFields     = []string{"Content", "Parent", "Username", "Timestamp"}
	newAllFields      = append(newOnlyFields, "Type", "UUID")
	metaOnlyFields    = []string{"Meta"}
	metaAllFields     = append(metaOnlyFields, "Type")
	queryAllFields    = []string{"Type", "UUID"}
)

// TestMarshalJSONWelcome ensures that the JSON serialization for WELCOME messages works
// and does not include message fields irrelevant for that message type.
func TestMarshalJSONWelcome(t *testing.T) {
	strWelcome := marshalOrFail(t, getWelcome())
	if strWelcome == "" {
		t.Error("Received empty string")
	}
	if contains(strWelcome, append(newOnlyFields, metaOnlyFields...), welcomeAllFields) {
		t.Error("Welcome message contained invalid fields", strWelcome)
	}
}

// TestMarshalJSONNew ensures that the JSON serialization for NEW messages works
// and does not include message fields irrelevant for that message type.
func TestMarshalJSONNew(t *testing.T) {
	strNewmessage := marshalOrFail(t, getNew())
	if strNewmessage == "" {
		t.Error("Received empty string")
	}
	if contains(strNewmessage, append(welcomeOnlyFields, metaOnlyFields...), newAllFields) {
		t.Error("New message contained invalid fields", strNewmessage)
	}
}

// TestMarshalJSONQuery ensures that the JSON serialization for QUERY messages works
// and does not include message fields irrelevant for that message type.
func TestMarshalJSONQuery(t *testing.T) {
	strQuery := marshalOrFail(t, getQuery())
	if strQuery == "" {
		t.Error("Received empty string")
	}
	if contains(strQuery, append(append(welcomeOnlyFields, newAllFields...), metaOnlyFields...), queryAllFields) {
		t.Error("Query message contained invalid fields", strQuery)
	}
}

// TestMarshalJSONMeta ensures that the JSON serialization for META messages works
// and does not include message fields irrelevant for that message type.
func TestMarshalJSONMeta(t *testing.T) {
	strMeta := marshalOrFail(t, getMeta())
	if strMeta == "" {
		t.Error("Received empty string")
	}
	if contains(strMeta, append(welcomeOnlyFields, newAllFields...), metaAllFields) {
		t.Error("Query message contained invalid fields", strMeta)
	}
}

// TestMarshalJSONInvalid ensures that the JSON serialization for invalid messages fails
func TestMarshalJSONInvalid(t *testing.T) {
	badOutput, err := getInvalid().MarshalJSON()
	if err == nil {
		t.Error("Marshalling bad message type should be error")
	}
	if badOutput != nil {
		t.Error("Marshalling bad message type should return nil slice, got", badOutput)
	}
}

// marshals the message into JSON, failing the test if the marshal returns an error.
func marshalOrFail(t *testing.T, m *arbor.ProtocolMessage) string {
	marhalled, err := m.MarshalJSON()
	if err != nil {
		t.Error(err)
	}
	return string(marhalled)
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

// TestRootIsValid ensures that our validation logic for messages takes the special properties
// of root messages (namely their lack of a parent id) into account.
func TestRootIsValid(t *testing.T) {
	rootMsg := &arbor.ProtocolMessage{
		Type: arbor.NewMessageType,
		ChatMessage: &arbor.ChatMessage{
			UUID:      "Something",
			Parent:    "",
			Username:  "thing",
			Content:   "Stuff",
			Timestamp: time.Now().Unix(),
		},
	}
	if !rootMsg.IsValid() {
		t.Error("Root message should be considered valid")
	}
}

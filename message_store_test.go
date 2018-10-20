package arbor_test

import (
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

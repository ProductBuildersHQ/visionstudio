package version

import "testing"

func TestStringNeverEmpty(t *testing.T) {
	got := String()
	if got == "" {
		t.Fatal("String() returned an empty string")
	}
}

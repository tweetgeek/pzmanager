package main

import "testing"

func TestVersion_HasDefault(t *testing.T) {
	if Version == "" {
		t.Error("Version should have a default value")
	}
}

package main

import (
	"slices"
	"testing"
)

func TestPathSet(t *testing.T) {
	pathSet := PathSet{}

	if len(pathSet) != 0 {
		t.Errorf("new path set should be empty; got lenght %d", len(pathSet))
	}

	pathSet.AddAll([]string{"new path"})
	if len(pathSet) != 1 {
		t.Error("fail add new entry")
	}

	expected := []string{"new path"}
	if entries := pathSet.Entries(); !slices.Equal(expected, entries) {
		t.Errorf("fail to generate entries: got %q; expected %q", entries, expected)
	}
}

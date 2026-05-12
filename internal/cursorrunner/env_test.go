package cursorrunner

import (
	"reflect"
	"testing"
)

func TestFilterEnv_removesForgePrefixedEntries(t *testing.T) {
	in := []string{
		"PATH=/usr/bin",
		"FORGE_FEATURE=11",
		"HOME=/Users/test",
		"FORGE_CURSOR_BIN=/opt/cursor/bin/cursor",
		"FORGE=keep-this-not-a-prefix-of-FORGE_",
	}
	got := FilterEnv(in)
	want := []string{
		"PATH=/usr/bin",
		"HOME=/Users/test",
		"FORGE=keep-this-not-a-prefix-of-FORGE_",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FilterEnv stripped wrong entries\n got: %v\nwant: %v", got, want)
	}
}

func TestFilterEnv_emptyInputReturnsEmpty(t *testing.T) {
	got := FilterEnv(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

package menu

import (
	"strings"
	"testing"
)

func TestReadKey(t *testing.T) {
	tests := map[string]key{
		" ":       keySelect,
		"\n":      keySelect,
		"\x1b":    keyEscape,
		"\x1b[A":  keyUp,
		"\x1b[B":  keyDown,
		"k":       keyUp,
		"j":       keyDown,
	}

	for input, want := range tests {
		got, err := readKey(strings.NewReader(input))
		if err != nil {
			t.Fatalf("read %q: %v", input, err)
		}
		if got != want {
			t.Fatalf("read %q = %v, want %v", input, got, want)
		}
	}
}

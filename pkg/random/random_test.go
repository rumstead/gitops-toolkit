package random

import (
	"strings"
	"testing"
)

func TestStringLength(t *testing.T) {
	for _, n := range []int{0, 1, 5, 32} {
		if got := String(n); len(got) != n {
			t.Errorf("String(%d) returned %q with length %d", n, got, len(got))
		}
	}
}

func TestStringCharset(t *testing.T) {
	s := String(256)
	for _, c := range s {
		if !strings.ContainsRune(letters, c) {
			t.Errorf("String returned unexpected character %q", c)
		}
	}
}

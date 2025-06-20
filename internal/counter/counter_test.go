package counter

import (
	"testing"
)

func TestCountWords(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"Hello world", 2},
		{"Go is awesome", 3},
		{"", 0},
		{"   ", 0},
		{"One\ntwo\nthree", 3},
		{"A quick brown fox jumps over the lazy dog", 9},
	}

	wc := WordCounter{}

	for _, test := range tests {
		result := wc.CountWords(test.input)
		if result != test.expected {
			t.Errorf("CountWords(%q) = %d; want %d", test.input, result, test.expected)
		}
	}
}
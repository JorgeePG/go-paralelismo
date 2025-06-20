package worker

import (
	"testing"
)

func TestProcess(t *testing.T) {
	worker := Worker{}

	tests := []struct {
		input    []string
		expected int
	}{
		{input: []string{"Hello world", "Go is great"}, expected: 5},
		{input: []string{"One line", "Another line"}, expected: 4},
		{input: []string{"", " "}, expected: 0},
		{input: []string{"Multiple words in a single line"}, expected: 7},
	}

	for _, test := range tests {
		result := worker.Process(test.input)
		if result != test.expected {
			t.Errorf("Process(%v) = %d; want %d", test.input, result, test.expected)
		}
	}
}
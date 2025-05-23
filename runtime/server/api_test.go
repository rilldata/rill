package server

import "testing"

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", ""},
		{"single word", "hello", "Hello"},
		{"multiple words", "hello world", "HelloWorld"},
		{"with underscores", "hello_world", "HelloWorld"},
		{"with dashes", "hello-world", "HelloWorld"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := toPascalCase(test.input)
			if result != test.expected {
				t.Errorf("expected %s, got %s", test.expected, result)
			}
		})
	}
}

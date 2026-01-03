package main

import (
	"reflect"
	"testing"
)

// 1. The "Table" - This is where you act like the Bot.
// Add every weird case you can think of here.
func TestDecodeBencode(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    interface{} // interface{} because result could be string, int, or list
		expectError bool        // whether we expect an error
	}{
		// STRING TESTS
		{"Basic String", "5:hello", "hello", false},
		{"Empty String", "0:", "", false},
		{"Long String", "12:hello world!", "hello world!", false},

		// INTEGER TESTS
		{"Basic Int", "i52e", 52, false},
		{"Negative Int", "i-42e", -42, false},
		{"Zero", "i0e", 0, false},

		// LIST TESTS (The hard part)
		{"Simple List", "l5:helloi52ee", []interface{}{"hello", 52}, false},
		{"Empty List", "le", []interface{}{}, false},
		{"Nested List", "ll5:helloee", []interface{}{[]interface{}{"hello"}}, false},

		// EDGE CASES
		{"Length 1 String", "1:a", "a", false},
		{"Zero String", "0:", "", false},
		{"Unterminated List", "l5:hello", nil, true},
	}

	// 2. The Loop - This runs your code against the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Replace 'decodeBencode' with whatever your function is named
			got, _, err := decodeBencode(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("For %s: expected an error, but got none", tt.name)
				}
				return // Skip the value check if we expected an error
			}

			if err != nil {
				t.Errorf("For %s, unexpected error: %v", tt.name, err)
			}

			// reflect.DeepEqual is needed to compare slices/lists in Go
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("For %s: expected %v, got %v", tt.name, tt.expected, got)
			}
		})
	}
}

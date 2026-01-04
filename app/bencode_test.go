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

		// DICT TESTS
		{"Simple Dict", "d3:foo3:bar5:helloi52ee", map[string]interface{}{"foo": "bar", "hello": 52}, false},
		{"Empty Dict", "de", map[string]interface{}{}, false},
		{"Dict with String Values", "d4:name4:John3:age6:twentye", map[string]interface{}{"name": "John", "age": "twenty"}, false},
		{"Dict with Int Values", "d3:agei25e6:heighti180ee", map[string]interface{}{"age": 25, "height": 180}, false},
		{"Dict with List Value", "d6:colorsli255ei0ei0eee", map[string]interface{}{"colors": []interface{}{255, 0, 0}}, false},
		{"Dict with Mixed Types", "d4:name3:Bob3:agei30e4:tagsl2:go4:codee6:activei1ee", map[string]interface{}{"name": "Bob", "age": 30, "tags": []interface{}{"go", "code"}, "active": 1}, false},
		{"Nested Dict", "d4:userd4:name4:Johne5:admini1ee", map[string]interface{}{"user": map[string]interface{}{"name": "John"}, "admin": 1}, false},
		{"Dict with Empty String Key", "d0:5:valuee", map[string]interface{}{"": "value"}, false},
		{"Dict with Empty String Value", "d3:key0:e", map[string]interface{}{"key": ""}, false},
		{"Dict with Negative Int", "d7:balancei-100ee", map[string]interface{}{"balance": -100}, false},
		{"Dict with Empty List Value", "d5:itemslee", map[string]interface{}{"items": []interface{}{}}, false},
		{"Dict with Empty Dict Value", "d8:metadatadee", map[string]interface{}{"metadata": map[string]interface{}{}}, false},

		// EDGE CASES
		{"Length 1 String", "1:a", "a", false},
		{"Negative Zero", "i-0e", 0, false},
		{"Zero String", "0:", "", false},
		{"Unterminated List", "l5:hello", nil, true},
		{"Unterminated Dict", "d5:hello", nil, true},
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

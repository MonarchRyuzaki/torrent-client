package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"unicode"
	// bencode "github.com/jackpal/bencode-go" // Available if you need it!
)

// Ensures gofmt doesn't remove the "os" encoding/json import (feel free to remove this!)
var _ = json.Marshal

func findFirstIndex(s string, r rune) int {
	var firstIdx int

	for i := 0; i < len(s); i++ {
		if rune(s[i]) == r {
			firstIdx = i
			break
		}
	}

	return firstIdx
}

// Example:
// - 5:hello -> hello
// - 10:hello12345 -> hello12345
// - i52e -> 52
// - i-52e -> -52
// - ["hello", 52] -> l 5:hello i52e e (no spaces)
func decodeBencode(bencodedString string) (interface{}, error) {
	if unicode.IsDigit(rune(bencodedString[0])) {
		firstColonIndex := findFirstIndex(bencodedString, rune(':'))

		lengthStr := bencodedString[:firstColonIndex]
		length, err := strconv.Atoi(lengthStr)
		if err != nil {
			return "", err
		}

		return bencodedString[firstColonIndex+1 : firstColonIndex+1+length], nil
	} else if rune(bencodedString[0]) == rune('i') {
		firstEIndex := findFirstIndex(bencodedString, rune('e'))

		num, err := strconv.Atoi(bencodedString[1:firstEIndex])
		if err != nil {
			return "", err
		}

		return num, nil
	} else if rune(bencodedString[0]) == rune('l') {
		decodedList := make([]interface{}, 0)
		ptr := 1
		for ptr < len(bencodedString) {
			if unicode.IsDigit(rune(bencodedString[ptr])) {
				str, err := decodeBencode(bencodedString[ptr:])
				if err != nil {
					return "", err
				}
				if x, ok := str.(string); !ok {
					return "", fmt.Errorf("%v is not a string", x)
				}
				strValue := str.(string)
				decodedList = append(decodedList, strValue)
				ptr += 1 + len(fmt.Sprint(len(strValue))) + len(strValue)
			} else if rune(bencodedString[ptr]) == rune('i') {
				num, err := decodeBencode(bencodedString[ptr:])
				if err != nil {
					return "", err
				}
				if x, ok := num.(int); !ok {
					return "", fmt.Errorf("%v is not a string", x)
				}
				value := num.(int)
				decodedList = append(decodedList, value)
				ptr += len(fmt.Sprint(value)) + 2
			} else if rune(bencodedString[ptr]) == rune('e') {
				ptr++
			} else {
				return "", fmt.Errorf("Only strings, integers are supported at the moment")
			}
		}
		return decodedList, nil
	} else {
		return "", fmt.Errorf("Only strings, integers, lists are supported at the moment")
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	command := os.Args[1]

	if command == "decode" {
		// TODO: Uncomment the code below to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	} else {
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

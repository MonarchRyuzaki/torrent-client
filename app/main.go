package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode"

	bencode "github.com/jackpal/bencode-go" // Available if you need it!
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

/*
	 Example:
		- i52e -> 52
		- i-52e -> -52
*/
func decodeInt(s string) (int, int, error) {
	firstEIndex := findFirstIndex(s, rune('e'))

	num, err := strconv.Atoi(s[1:firstEIndex])
	if err != nil {
		return 0, 0, err
	}
	len := firstEIndex + 1
	return num, len, nil
}

/*
	 Example:
		- 5:hello -> hello
		- 10:hello12345 -> hello12345
*/
func decodeString(s string) (string, int, error) {
	firstColonIndex := findFirstIndex(s, rune(':'))

	lengthStr := s[:firstColonIndex]
	length, err := strconv.Atoi(lengthStr)
	if err != nil {
		return "", 0, err
	}
	len := len(lengthStr) + length + 1
	return s[firstColonIndex+1 : firstColonIndex+1+length], len, nil
}

/*
	 Example:
		- ["hello", 52] -> l 5:hello i52e e (no spaces)
*/
func decodeList(bencodedString string) ([]interface{}, int, error) {
	decodedList := make([]interface{}, 0)
	ptr := 1
	for ptr < len(bencodedString) {
		// fmt.Println(ptr, len(bencodedString))
		if bencodedString[ptr] == 'e' {
			return decodedList, ptr + 1, nil
		}
		item, lengthProcessed, err := decodeBencode(bencodedString[ptr:])
		if err != nil {
			return nil, 0, err
		}
		decodedList = append(decodedList, item)
		ptr += lengthProcessed
	}
	return nil, 0, fmt.Errorf("unterminated list")
}

/*
	 Example:
		- d3:foo3:bar5:helloi52ee -> {"foo":"bar","hello":52}
*/
func decodeDict(bencodedString string) (map[string]interface{}, int, error) {
	decodedMap := make(map[string]interface{})
	ptr := 1
	key := "!@#"
	for ptr < len(bencodedString) {
		if bencodedString[ptr] == 'e' {
			return decodedMap, ptr + 1, nil
		}
		item, lengthProcessed, err := decodeBencode(bencodedString[ptr:])
		if err != nil {
			return nil, 0, err
		}
		if key == "!@#" {
			if _, ok := item.(string); !ok {
				return nil, 0, fmt.Errorf("Key is not a string")
			}
			key = item.(string)
		} else {
			decodedMap[key] = item
			key = "!@#"
		}
		ptr += lengthProcessed
	}
	return nil, 0, fmt.Errorf("unterminated dict")
}

func decodeBencode(bencodedString string) (interface{}, int, error) {
	// fmt.Println(bencodedString)
	if unicode.IsDigit(rune(bencodedString[0])) {
		str, len, err := decodeString(bencodedString)

		if err != nil {
			return "", 0, err
		}

		return str, len, err
	} else if rune(bencodedString[0]) == rune('i') {
		num, len, err := decodeInt(bencodedString)

		if err != nil {
			return "", 0, err
		}

		return num, len, err
	} else if rune(bencodedString[0]) == rune('l') {
		list, len, err := decodeList(bencodedString)

		if err != nil {
			return "", 0, err
		}

		return list, len, err
	} else if rune(bencodedString[0]) == rune('d') {
		mpp, len, err := decodeDict(bencodedString)

		if err != nil {
			return "", 0, err
		}

		return mpp, len, err
	} else {
		return "", 0, fmt.Errorf("Only strings, integers, lists and dictionary are supported at the moment")
	}
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Fprintln(os.Stderr, "Logs from your program will appear here!")

	command := os.Args[1]

	switch command {
	case "decode":
		// TODO: Uncomment the code below to pass the first stage
		//
		bencodedValue := os.Args[2]

		decoded, _, err := decodeBencode(bencodedValue)
		if err != nil {
			fmt.Println(err)
			return
		}

		jsonOutput, _ := json.Marshal(decoded)
		fmt.Println(string(jsonOutput))
	case "info":
		torrentFile := os.Args[2]
		bencodedValue, err := os.ReadFile(torrentFile)
		if err != nil {
			fmt.Println(err)
			return
		}
		torrentInfo := BencodeFile{}
		e := bencode.Unmarshal(strings.NewReader(string(bencodedValue)), &torrentInfo)
		if e != nil {
			fmt.Println(e)
			return
		}
		jsonOutput, _ := json.Marshal(torrentInfo)
		fmt.Println(string(jsonOutput))
	default:
		fmt.Println("Unknown command: " + command)
		os.Exit(1)
	}
}

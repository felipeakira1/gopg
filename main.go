package main

import (
	"encoding/binary"
	"fmt"
	"strings"
)

func startupMessage(user, database string) ([]byte, error) {
	// Checks for empty user or database strings
	if user == "" || database == "" {
		return nil, fmt.Errorf("user and database are required")
	}

	// Checks for null bytes in user or database strings
	// \x00 represents a null byte, which is used to terminate strings
	if strings.ContainsRune(user, '\x00') || strings.ContainsRune(database, '\x00') {
		return nil, fmt.Errorf("user and database cannot contain null bytes")
	}

	params := []byte(
		"user\x00" + user + "\x00" +
			"database\x00" + database + "\x00" +
			"\x00",
	)

	message := make([]byte, 8+len(params))

	// binary from the enconding/binary package
	// binary.BigEndian.PutUint32 writes a uint32 value to a byte slice
	// message[0:4] selects the first four bytes
	binary.BigEndian.PutUint32(message[0:4], uint32(len(message)))

	// 3 << 16 places 3 in the upper 16 bits, giving 196608
	version := uint32(3) << 16
	binary.BigEndian.PutUint32(message[4:8], version)

	copy(message[8:], params)

	return message, nil
}

func main() {}

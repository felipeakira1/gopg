package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"strings"
	"time"
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

func connect(user, database string) (net.Conn, error) {
	// Create the startup message
	message, err := startupMessage(user, database)

	if err != nil {
		return nil, err
	}

	fmt.Println("Establishing connection to PostgreSQL at ", time.Now().Format("02/01/2006 15:04:05.000"))
	// Establish a TCP connection to the PostgreSQL server
	conn, err := net.DialTimeout("tcp", "127.0.0.1:15432", 5*time.Second)
	if err != nil {
		return nil, err
	}

	// Set a deadline for the connection to avoid hanging indefinitely
	// if err := conn.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
	// 	conn.Close()
	// 	return nil, err
	// 	}

	// bytes package provides functions for the manipulation of bytes slices
	// bytes.NewReader creates a new Reader reading from the provided byte slice
	// io from the io package provides basic interfaces to I/O primitives
	// io.Copy reads from that reader and writes to conn
	// _ discards the returned number of bytes copied
	if _, err := io.Copy(conn, bytes.NewReader(message)); err != nil {
		conn.Close()
		return nil, err
	}

	return conn, nil
}

func main() {}

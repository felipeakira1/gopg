package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"log"
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

func readMessage(conn net.Conn) (byte, []byte, error) {
	header := make([]byte, 5)

	if _, err := io.ReadFull(conn, header); err != nil {
		return 0, nil, fmt.Errorf("read header: %w", err)
	}

	messageType := header[0]
	length := binary.BigEndian.Uint32(header[1:5])

	if length < 4 || length > 16*1024*1024 {
		return 0, nil, fmt.Errorf("invalid message length: %d", length)
	}

	body := make([]byte, length-4)

	if _, err := io.ReadFull(conn, body); err != nil {
		return 0, nil, fmt.Errorf("read body: %w", err)
	}

	return messageType, body, nil
}

func receiveUntilReady(conn net.Conn) error {
	for {
		messageType, body, err := readMessage(conn)
		if err != nil {
			return err
		}

		switch messageType {
		case 'R':
			if len(body) < 4 {
				log.Println("Invalid authentication message length")
				return fmt.Errorf("invalid authentication message")
			}
			authCode := binary.BigEndian.Uint32(body[:4])

			if authCode != 0 {
				return fmt.Errorf("unsupported authentication method: %d", authCode)
			}

			if len(body) != 4 {
				return fmt.Errorf("invalid AuthenticationOk length")
			}

			fmt.Println("AuthenticationOK received")
		case 'S':
			fmt.Printf("ParameterStatus received: %s\n", string(body))
		case 'K':
			fmt.Println("Received BackendKeyData")
		case 'Z':
			fmt.Println("ReadyForQuery received")
			return nil

		case 'C':
			fmt.Println("Query command: ", string(body))

		default:
			fmt.Printf("Message: %c, body: % x\n", messageType, body)
			text := string(body)
			fmt.Println(text)
		}
	}
}

type FieldDescription struct {
	Name         string
	TableOID     uint32
	ColumnNumber uint16
	TypeOID      uint32
	TypeSize     uint16
	TypeModifier uint32
	FormatCode   uint16
}

func decodeRowDescription(body []byte) ([]FieldDescription, error) {
	// The first two bytes are required and represent the number of fields in a row.
	if len(body) < 2 {
		return nil, fmt.Errorf("missing field count")
	}

	numberOfFields := binary.BigEndian.Uint16(body[0:2])
	fields := make([]FieldDescription, 0, numberOfFields)
	offset := 2

	for i := 0; i < int(numberOfFields); i++ {
		end := bytes.IndexByte(body[offset:], 0)
		if end == -1 {
			return nil, fmt.Errorf("missing field name terminator")
		}

		name := string(body[offset : offset+end])
		offset += end + 1

		if len(body)-offset < 18 {
			return nil, fmt.Errorf("incomplete field metadata")
		}

		rest := body[offset : offset+18]

		field := FieldDescription{
			Name:         name,
			TableOID:     binary.BigEndian.Uint32(rest[0:4]),
			ColumnNumber: binary.BigEndian.Uint16(rest[4:6]),
			TypeOID:      binary.BigEndian.Uint32(rest[6:10]),
			TypeSize:     binary.BigEndian.Uint16(rest[10:12]),
			TypeModifier: binary.BigEndian.Uint32(rest[12:16]),
			FormatCode:   binary.BigEndian.Uint16(rest[16:18]),
		}

		fields = append(fields, field)

		offset += 18
	}

	if offset != len(body) {
		return nil, fmt.Errorf("unexpected trailing bytes")
	}

	return fields, nil
}

func main() {}

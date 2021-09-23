package message

import (
	"encoding/binary"
	"errors"
	"io"
	"math"
)

// Encode encode message
func Encode(msg string) ([]byte, error) {
	content := []byte(msg)
	if len(content) > math.MaxUint16 {
		return []byte{}, errors.New("message is to big")
	}
	res := make([]byte, 2, len(content)+2)
	binary.BigEndian.PutUint16(res, uint16(len(content)))
	res = append(res, content...)
	return res, nil
}

// Decode decode message
func Decode(data []byte) (string, error) {
	formatError := errors.New("wrong message format")
	if cap(data) < 2 {
		return "", formatError
	}
	cLen := binary.BigEndian.Uint16(data[0:2])
	if cLen != uint16(len(data[2:])) {
		return "", formatError
	}

	return string(data[2:]), nil
}

// Read reads message from io.Reader
func Read(r io.Reader) ([]byte, error) {
	lb := make([]byte, 2)
	_, err := io.ReadFull(r, lb)
	if err != nil {
		return nil, err
	}
	content := make([]byte, binary.BigEndian.Uint16(lb))
	_, err = io.ReadFull(r, content)
	if err != nil {
		return nil, err
	}
	return append(lb, content...), nil
}

package menu

import (
	"errors"
	"io"
)

type key int

const (
	keyUnknown key = iota
	keyUp
	keyDown
	keySelect
	keyEscape
)

func readKey(reader io.Reader) (key, error) {
	for {
		first, ok, err := readByte(reader)
		if errors.Is(err, io.EOF) {
			continue
		}
		if err != nil {
			return keyUnknown, err
		}
		if !ok {
			continue
		}

		switch first {
		case ' ', '\r', '\n':
			return keySelect, nil
		case 'k':
			return keyUp, nil
		case 'j':
			return keyDown, nil
		case 0x1b:
			second, ok, err := readByte(reader)
			if errors.Is(err, io.EOF) {
				return keyEscape, nil
			}
			if err != nil {
				return keyUnknown, err
			}
			if !ok {
				return keyEscape, nil
			}
			if second != '[' {
				return keyEscape, nil
			}
			third, ok, err := readByte(reader)
			if errors.Is(err, io.EOF) {
				return keyEscape, nil
			}
			if err != nil {
				return keyUnknown, err
			}
			if !ok {
				return keyEscape, nil
			}
			switch third {
			case 'A':
				return keyUp, nil
			case 'B':
				return keyDown, nil
			default:
				return keyUnknown, nil
			}
		}
	}
}

func readByte(reader io.Reader) (byte, bool, error) {
	var buffer [1]byte
	n, err := reader.Read(buffer[:])
	if n == 1 {
		return buffer[0], true, nil
	}
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, false, io.EOF
		}
		return 0, false, err
	}
	return 0, false, nil
}

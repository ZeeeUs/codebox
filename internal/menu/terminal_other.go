//go:build !linux && !darwin

package menu

import (
	"errors"
	"io"
)

func makeRaw(io.Reader) (func(), error) {
	return nil, errors.New("interactive menu is not supported on this operating system")
}

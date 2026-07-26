//go:build linux

package menu

import (
	"io"
	"os"
	"syscall"
	"unsafe"
)

func makeRaw(reader io.Reader) (func(), error) {
	file, ok := reader.(*os.File)
	if !ok {
		return func() {}, nil
	}

	fd := file.Fd()
	var original syscall.Termios
	if _, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(&original)),
		0,
		0,
		0,
	); errno != 0 {
		return nil, errno
	}

	raw := original
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Cflag |= syscall.CS8
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 0
	raw.Cc[syscall.VTIME] = 1

	if err := setTermios(fd, &raw); err != nil {
		return nil, err
	}

	return func() {
		_ = setTermios(fd, &original)
	}, nil
}

func setTermios(fd uintptr, state *syscall.Termios) error {
	_, _, errno := syscall.Syscall6(
		syscall.SYS_IOCTL,
		fd,
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(state)),
		0,
		0,
		0,
	)
	if errno != 0 {
		return errno
	}
	return nil
}

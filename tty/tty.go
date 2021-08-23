package tty

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

type TTY struct {
	Command string
	Args    []string
	Buffer  *CircleBuffer
	cmd     *exec.Cmd
	PTY     *os.File
}

func (t *TTY) Write(data []byte) error {

	_, err := t.PTY.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func (t *TTY) Read() ([]byte, error) {

	data := make([]byte, 512)
	_, err := t.PTY.Read(data)
	if err != nil {
		return nil, err
	}

	_, err = t.Buffer.Write(data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (t *TTY) Resize(width, height int) error {

	window := struct {
		row uint16
		col uint16
		x   uint16
		y   uint16
	}{
		uint16(height),
		uint16(width),
		0,
		0,
	}
	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		t.PTY.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&window)),
	)
	if errno != 0 {
		return errno
	}
	return nil
}

func (t *TTY) Close() {
	t.PTY.Close()
}

func New(command string, argv []string) (*TTY, error) {

	cmd := exec.Command(command, argv...)

	pty, err := pty.Start(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to start pty: %s", err.Error)
	}

	newTTY := &TTY{
		Command: command,
		Args:    argv,
		cmd:     cmd,
		PTY:     pty,
	}
	newTTY.Buffer = NewCircleBuffer(1024 * 1024)
	return newTTY, nil
}

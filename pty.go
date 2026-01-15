package main

import (
	"os"
	"os/exec"

	cpty "github.com/creack/pty"
)

type pty struct {
	cmd *exec.Cmd
	tty *os.File
}

func newPty(name string, args ...string) (pty, error) {
	cmd := exec.Command(name, args...)

	tty, err := cpty.StartWithSize(cmd, &cpty.Winsize{
		Rows: uint16(emulatorRows),
		Cols: uint16(emulatorCols),
	})

	if err != nil {
		return pty{}, err
	}

	return pty{
		cmd,
		tty,
	}, nil
}

func (p *pty) write(input []byte) {
	if p.tty == nil {
		return
	}

	p.tty.Write(input)
}

func (p *pty) read(output []byte) (int, error) {
	return p.tty.Read(output)
}

package main

import (
	"os"
	"os/exec"

	cpty "github.com/creack/pty"
)

type pty struct {
	tty *os.File
}

func newPty(cmd *exec.Cmd) (pty, error) {
	tty, err := cpty.StartWithAttrs(cmd, &cpty.Winsize{
		Rows: uint16(emulatorRows),
		Cols: uint16(emulatorCols),
	}, nil)

	if err != nil {
		return pty{}, err
	}

	return pty{tty}, nil
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

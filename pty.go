package main

import (
	"os"
	"os/exec"

	cpty "github.com/creack/pty"
)

type pty struct {
	tty *os.File
}

func newPty(cmd *exec.Cmd, rows, cols int) (pty, error) {
	tty, err := cpty.StartWithAttrs(cmd, &cpty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	}, nil)

	if err != nil {
		return pty{}, err
	}

	return pty{tty}, nil
}

func (p *pty) Resize(rows, cols int) error {
	if p.tty == nil {
		return nil
	}

	return cpty.Setsize(p.tty, &cpty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
}

func (p *pty) write(input []byte) error {
	if p.tty == nil {
		return nil
	}

	_, err := p.tty.Write(input)

	return err
}

func (p *pty) read(output []byte) (int, error) {
	return p.tty.Read(output)
}

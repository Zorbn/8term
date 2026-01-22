package main

import (
	"os"
	"os/exec"
	"syscall"

	cpty "github.com/creack/pty"
)

type pty struct {
	tty *os.File
}

func newPty(cmd *exec.Cmd, rows, cols int) (pty, error) {
	tty, err := start(cmd, &cpty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})

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

// A copy of creack/pty StartWithAttrs but with Ctty and ExtraFiles set up to allow stdin piping.
func start(c *exec.Cmd, sz *cpty.Winsize) (*os.File, error) {
	pty, tty, err := cpty.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tty.Close() }() // Best effort.

	if sz != nil {
		if err := cpty.Setsize(pty, sz); err != nil {
			_ = pty.Close() // Best effort.
			return nil, err
		}
	}
	if c.Stdout == nil {
		c.Stdout = tty
	}
	if c.Stderr == nil {
		c.Stderr = tty
	}
	if c.Stdin == nil {
		c.Stdin = tty
	}

	c.ExtraFiles = []*os.File{tty}

	c.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    3,
	}

	if err := c.Start(); err != nil {
		_ = pty.Close() // Best effort.
		return nil, err
	}
	return pty, err
}

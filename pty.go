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

func newPty(name string, arg ...string) pty {
	cmd := exec.Command(name, arg...)
	tty, err := cpty.Start(cmd)

	cpty.Setsize(tty, &cpty.Winsize{
		Rows: uint16(emulatorRows),
		Cols: uint16(emulatorCols),
	})

	if err != nil {
		panic(err)
	}

	return pty{
		cmd,
		tty,
	}
}

func (p *pty) write(input []byte) {
	p.tty.Write(input)
}

func (p *pty) read(output []byte) (int, error) {
	return p.tty.Read(output)
}

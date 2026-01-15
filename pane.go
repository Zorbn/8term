package main

import (
	_ "embed"
	"fmt"
	"io"
	"log"

	"github.com/danielgatis/go-vte"
)

type pane struct {
	pty      pty
	buffer   []byte
	output   chan []byte
	parser   *vte.Parser
	emulator emulator
}

func newPane() pane {
	buffer := make([]byte, 4096)
	output := make(chan []byte)
	emulator := newEmulator()

	return pane{
		pty{},
		buffer,
		output,
		nil,
		emulator,
	}
}

func (p *pane) run(ast astNode) error {
	if p.parser == nil {
		p.parser = vte.NewParser(&p.emulator)
	}

	go func() {
		ast.exec(p)
		close(p.output)
	}()

	return nil
}

func (p *pane) runToExit(name string, args ...string) int {
	var err error
	p.pty, err = newPty(name, args...)

	if err != nil {
		p.output <- fmt.Appendf([]byte{}, "\x1b[0;31mUnable to run program: %s\x1b[m", name)
		return 1
	}

	for {
		outputLen, err := p.pty.read(p.buffer)

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatal(err)
		}

		// TODO: Use sync pool to avoid excess allocations.
		output := make([]byte, outputLen)
		copy(output, p.buffer[:outputLen])

		p.output <- output
	}

	p.pty.tty.Close()
	p.pty.cmd.Wait()

	return p.pty.cmd.ProcessState.ExitCode()
}

func (p *pane) handleOutput() bool {
loop:
	for {
		select {
		case output, ok := <-p.output:
			if !ok {
				return false
			}

			for _, b := range output {
				p.parser.Advance(b)
			}
		default:
			break loop
		}
	}

	input := p.emulator.input.Bytes()
	p.emulator.input.Reset()
	p.pty.write(input)

	return true
}

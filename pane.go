package main

import (
	_ "embed"
	"io"
	"log"
	"os/exec"

	"github.com/danielgatis/go-vte"
)

type pane struct {
	pty      pty
	buffer   []byte
	output   chan []byte
	parser   *vte.Parser
	emulator emulator
	exitCode int
	timer    float32
}

func newPane() pane {
	buffer := make([]byte, 4096)
	output := make(chan []byte)

	return pane{
		pty{},
		buffer,
		output,
		nil,
		emulator{},
		0,
		0,
	}
}

func (p *pane) run(ast astNode) error {
	go func() {
		p.exitCode = ast.exec(p)
		close(p.output)
	}()

	return nil
}

func (p *pane) runToExit(cmd *exec.Cmd) error {
	var err error
	p.pty, err = newPty(cmd)

	if err != nil {
		return err
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

	return nil
}

func (p *pane) handleOutput() bool {
	didAdvance := false

loop:
	for {
		select {
		case output, ok := <-p.output:
			if !ok {
				return false
			}

			if p.parser == nil {
				p.emulator = newEmulator()
				p.parser = vte.NewParser(&p.emulator)
			}

			for _, b := range output {
				p.parser.Advance(b)
			}

			didAdvance = false
		default:
			break loop
		}
	}

	if didAdvance {
		input := p.emulator.input.Bytes()
		p.emulator.input.Reset()
		p.pty.write(input)
	}

	return true
}

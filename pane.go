package main

import (
	_ "embed"
	"io"
	"log"
)

type pane struct {
	pty      pty
	buffer   []byte
	output   chan []byte
	emulator emulator
}

func newPane(name string, arg ...string) (pane, error) {
	pty, err := newPty(name, arg...)

	if err != nil {
		return pane{}, err
	}

	buffer := make([]byte, 4096)
	output := make(chan []byte)
	emulator := newEmulator()

	return pane{
		pty,
		buffer,
		output,
		emulator,
	}, nil
}

func (p *pane) run() {
	go func() {
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
	}()
}

func (p *pane) handleOutput() {
	select {
	case output := <-p.output:
		ParseEscapeSequences(output, &p.emulator)
	default:
	}

	input := p.emulator.input.Bytes()
	p.emulator.input.Reset()
	p.pty.write(input)
}

package main

import (
	_ "embed"
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

func newPane(name string, arg ...string) (pane, error) {
	pty, err := newPty(name, arg...)

	if err != nil {
		return pane{}, err
	}

	buffer := make([]byte, 4096)
	output := make(chan []byte)
	emulator := newEmulator()
	var parser *vte.Parser

	return pane{
		pty,
		buffer,
		output,
		parser,
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
	if p.parser == nil {
		p.parser = vte.NewParser(&p.emulator)
	}

loop:
	for {
		select {
		case output := <-p.output:
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
}

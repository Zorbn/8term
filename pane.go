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

func newPane(name string, arg ...string) pane {
	pty := newPty(name, arg...)
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
	}
}

func (p *pane) run() {
	p.parser = vte.NewParser(&p.emulator)

	go func() {
		for {
			outputLen, err := p.pty.read(p.buffer)

			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			p.output <- p.buffer[:outputLen]
		}

		p.pty.tty.Close()
	}()
}

func (p *pane) handleOutput() {
	select {
	case output := <-p.output:
		for _, b := range output {
			p.parser.Advance(b)
		}
	default:
	}
}

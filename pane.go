package main

import (
	_ "embed"
	"io"
	"log"
	"os/exec"

	"github.com/danielgatis/go-vte"
)

const bufferLen = 1024
const bufferCount = 8

type pane struct {
	rows, cols int
	pty        pty
	buffer     []byte
	output     chan []byte
	parser     *vte.Parser
	emulator   emulator
	exitCode   int
	timer      float32
}

func newPane(rows, cols int) pane {
	buffer := make([]byte, bufferLen)
	output := make(chan []byte, bufferCount)

	return pane{
		rows, cols,
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
	p.pty, err = newPty(cmd, p.rows, p.cols)

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
				p.emulator = newEmulator(p.rows, p.cols)
				p.parser = vte.NewParser(&p.emulator)
			}

			for _, b := range output {
				p.parser.Advance(b)
			}

			didAdvance = true
		default:
			break loop
		}
	}

	if didAdvance && p.emulator.input.Len() > 0 {
		input := p.emulator.input.Bytes()
		p.emulator.input.Reset()
		p.pty.write(input)
	}

	return true
}

func (p *pane) Resize(rows, cols int) {
	p.rows = rows
	p.cols = cols

	if p.parser != nil {
		p.emulator.Resize(rows, cols)
	}

	p.pty.Resize(rows, cols)
}

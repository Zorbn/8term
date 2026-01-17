package main

import (
	"fmt"
	"os"
	"os/exec"
)

func (c *callNode) exec(pane *pane) int {
	process := runCall(c, pane, nil, nil)
	process.wait()

	return process.exitCode
}

func (b *binaryNode) exec(pane *pane) int {
	switch b.op {
	case ";":
		b.left.exec(pane)

		return b.right.exec(pane)
	case "&&":
		leftExitCode := b.left.exec(pane)

		if leftExitCode != 0 {
			return leftExitCode
		}

		return b.right.exec(pane)
	case "||":
		leftExitCode := b.left.exec(pane)

		if leftExitCode == 0 {
			return leftExitCode
		}

		return b.right.exec(pane)
	default:
		panic("Unknown binary op")
	}
}

func (p *pipeNode) exec(pane *pane) int {
	var processes []process
	var input *os.File
	var output *os.File

	for i, call := range p.children {
		shouldPipe := i < len(p.children)-1
		var nextInput *os.File

		if shouldPipe {
			var err error
			nextInput, output, err = os.Pipe()

			if err != nil {
				panic(err)
			}
		}

		process := runCall(call, pane, input, output)
		processes = append(processes, process)

		if input != nil {
			input.Close()
			input = nil
		}

		if output != nil {
			output.Close()
			output = nil
		}

		if process.exitCode != 0 {
			nextInput.Close()
			break
		}

		input = nextInput
	}

	for _, process := range processes {
		process.wait()
	}

	lastProcess := processes[len(processes)-1]

	return lastProcess.exitCode
}

func runCall(call *callNode, pane *pane, input *os.File, output *os.File) process {
	name := call.children[0]
	args := call.children[1:]

	if process, isBuiltin := tryRunBuiltin(name, args, pane, input, output); isBuiltin {
		return process
	}

	cmd := exec.Command(name, args...)

	cmd.Stdin = input

	var err error

	if output != nil {
		cmd.Stdout = output
		err = cmd.Start()
	} else {
		err = pane.runToExit(cmd)
	}

	if err != nil {
		writeCallErrorToPane(name, pane)
		return process{exitCode: 1}
	}

	return process{cmd: cmd}
}

func tryRunBuiltin(name string, args []string, pane *pane, input *os.File, output *os.File) (process, bool) {
	switch name {
	case "cd", "help":
	default:
		return process{}, false
	}

	switch name {
	case "cd":
		if len(args) > 1 {
			return process{exitCode: 1}, true
		}

		var path string

		if len(args) == 1 {
			path = args[0]
		} else {
			var err error
			path, err = os.UserHomeDir()

			if err != nil {
				return process{exitCode: 1}, true
			}
		}

		os.Chdir(path)
	case "help":
		if len(args) != 0 {
			return process{exitCode: 1}, true
		}

		writeFromBuiltin("Try 'cd' to change directories.", pane, output)
	}

	return process{exitCode: 0}, true
}

func writeFromBuiltin(text string, pane *pane, output *os.File) {
	textBytes := []byte(text)

	if output != nil {
		output.Write(textBytes)
	} else {
		pane.output <- textBytes
	}
}

func writeCallErrorToPane(name string, pane *pane) {
	pane.output <- fmt.Appendf([]byte{}, "\x1b[0;31mUnable to run program: %s\x1b[m\r\n", name)
}

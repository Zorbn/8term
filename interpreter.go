package main

import (
	"fmt"
	"os"
	"os/exec"
)

func (c *callNode) exec(pane *pane) int {
	cmd, exitCode := runCall(c, pane, nil, nil)

	if exitCode != 0 {
		return exitCode
	}

	if cmd != nil {
		cmd.Wait()

		return cmd.ProcessState.ExitCode()
	}

	return 0
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
	var commands []*exec.Cmd
	var input *os.File
	var output *os.File
	var exitCode int

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

		var cmd *exec.Cmd

		cmd, exitCode = runCall(call, pane, input, output)

		if input != nil {
			input.Close()
			input = nil
		}

		if output != nil {
			output.Close()
			output = nil
		}

		if exitCode != 0 {
			nextInput.Close()
			break
		}

		input = nextInput

		if cmd != nil {
			commands = append(commands, cmd)
		}
	}

	for _, cmd := range commands {
		cmd.Wait()
	}

	if exitCode != 0 {
		return exitCode
	}

	lastCommand := commands[len(commands)-1]

	return lastCommand.ProcessState.ExitCode()
}

func runCall(call *callNode, pane *pane, input *os.File, output *os.File) (*exec.Cmd, int) {
	name := call.children[0]
	args := call.children[1:]

	if isBuiltin, exitCode := tryRunBuiltin(name, args, pane, input, output); isBuiltin {
		return nil, exitCode
	}

	cmd := exec.Command(name, args...)

	cmd.Stdin = input

	var err error

	if output != nil {
		fmt.Println(name, "cmd start")
		cmd.Stdout = output
		err = cmd.Start()
	} else {
		fmt.Println(name, "pane start")
		err = pane.runToExit(cmd)
	}

	if err != nil {
		writeCallErrorToPane(name, pane)
		return nil, 1
	}

	return cmd, 0
}

func tryRunBuiltin(name string, args []string, pane *pane, input *os.File, output *os.File) (isBuiltin bool, exitCode int) {
	switch name {
	case "cd", "help":
		isBuiltin = true
	default:
		return
	}

	switch name {
	case "cd":
		if len(args) > 1 {
			exitCode = 1
			return
		}

		var path string

		if len(args) == 1 {
			path = args[0]
		} else {
			var err error
			path, err = os.UserHomeDir()

			if err != nil {
				exitCode = 1
				return
			}
		}

		os.Chdir(path)
	case "help":
		if len(args) != 0 {
			exitCode = 1
			return
		}

		writeFromBuiltin("Try 'cd' to change directories.", pane, output)
	}

	return
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
	pane.output <- fmt.Appendf([]byte{}, "\x1b[0;31mUnable to run program: %s\x1b[m", name)
}

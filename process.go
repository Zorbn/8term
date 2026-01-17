package main

import "os/exec"

type process struct {
	cmd      *exec.Cmd
	exitCode int
}

func (p *process) wait() {
	if p.cmd == nil {
		return
	}

	p.cmd.Wait()
	p.exitCode = p.cmd.ProcessState.ExitCode()
}

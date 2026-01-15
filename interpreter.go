package main

func (c *callNode) exec(pane *pane) int {
	name := string(c.tokens[0])
	args := make([]string, 0, len(c.tokens[1:]))

	for _, arg := range c.tokens[1:] {
		args = append(args, string(arg))
	}

	return pane.runToExit(name, args...)
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
	panic("TODO: Pipe is unimplemented")
}

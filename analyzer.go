package main

func (c *callNode) analyze(command *command) {
	name := c.children[0]
	command.historicalUsages[name]++
}

func (b *binaryNode) analyze(command *command) {
	b.left.analyze(command)
	b.right.analyze(command)
}

func (p *pipeNode) analyze(command *command) {
	for _, child := range p.children {
		child.analyze(command)
	}
}

package main

func (c *callNode) find(index int) astNode {
	if !nodeContains(c, index) {
		return nil
	}

	return c
}

func (b *binaryNode) find(index int) astNode {
	if !nodeContains(b, index) {
		return nil
	}

	if result := b.left.find(index); result != nil {
		return result
	}

	if result := b.right.find(index); result != nil {
		return result
	}

	return b
}

func (p *pipeNode) find(index int) astNode {
	if !nodeContains(p, index) {
		return nil
	}

	for _, child := range p.children {
		if result := child.find(index); result != nil {
			return result
		}
	}

	return p
}

func nodeContains(node astNode, index int) bool {
	return node.Start() <= index && index <= node.End()
}

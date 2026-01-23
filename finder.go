package main

func (c *callNode) find(index int) astNode {
	return c
}

func (b *binaryNode) find(index int) astNode {
	if result := b.left.find(index); result != nil {
		return result
	}

	if result := b.right.find(index); result != nil {
		return result
	}

	return b
}

func (p *pipeNode) find(index int) astNode {
	for _, child := range p.children {
		if result := child.find(index); result != nil {
			return result
		}
	}

	return p
}

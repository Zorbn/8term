package main

type command struct {
	runes     []rune
	tokenizer tokenizer
	parser    parser
	isDirty   bool
}

func (c *command) append(r rune) {
	c.runes = append(c.runes, r)
	c.isDirty = true
}

func (c *command) pop() {
	if len(c.runes) == 0 {
		return
	}

	c.runes = c.runes[:len(c.runes)-1]
	c.isDirty = true
}

func (c *command) clear() {
	if len(c.runes) == 0 {
		return
	}

	c.runes = c.runes[:0]
	c.isDirty = true
}

func (c *command) parse() (astNode, bool) {
	if c.isDirty {
		c.isDirty = false

		c.tokenizer.tokenize(c.runes)
		c.parser.parse(c.tokenizer.tokens)
	}

	return c.parser.ast, c.tokenizer.didSucceed && len(c.parser.errors) == 0
}

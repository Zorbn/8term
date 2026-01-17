package main

import (
	"os"
	"path/filepath"
	"strings"
)

type command struct {
	runes      []rune
	tokenizer  tokenizer
	parser     parser
	isDirty    bool
	completion []rune
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
		c.updateCompletion()
	}

	return c.parser.ast, c.tokenizer.didSucceed && len(c.parser.errors) == 0
}

func (c *command) updateCompletion() {
	c.parse()
	c.completion = c.completion[:0]

	call := c.parser.lastCallNode

	if call == nil || len(call.children) == 0 {
		return
	}

	path := call.children[len(call.children)-1]
	dir, file := filepath.Split(path)

	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	match := ""
	isDir := false

	for _, entry := range entries {
		name := entry.Name()

		if strings.HasPrefix(name, file) {
			match = name
			isDir = entry.IsDir()
			break
		}
	}

	if len(match) < len(file) {
		return
	}

	for _, r := range match[len(file):] {
		c.completion = append(c.completion, r)
	}

	if isDir {
		c.completion = append(c.completion, '/')
	}
}

func (c *command) applyCompletion() {
	for _, r := range c.completion {
		c.append(r)
	}

	c.completion = c.completion[:0]
}

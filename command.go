package main

import (
	"os"
	"path/filepath"
	"strings"
)

type command struct {
	runes           []rune
	tokenizer       tokenizer
	parser          parser
	isDirty         bool
	completion      []rune
	pathExecutables []string
}

func newCommand() command {
	pathExecutables := getPathExecutables()

	return command{pathExecutables: pathExecutables}
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

	match := ""

	if len(call.children) == 1 && c.runes[len(c.runes)-1] != ' ' {
		prefix := call.children[0]

		for _, executable := range c.pathExecutables {
			if (match == "" || len(executable) < len(match)) && strings.HasPrefix(executable, prefix) {

				match = executable
			}
		}
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

	isDir := false

	for _, entry := range entries {
		name := entry.Name()

		if (match == "" || len(name) < len(match)) && strings.HasPrefix(name, file) {
			match = name
			isDir = entry.IsDir()
		}
	}

	if len(match) < len(file) {
		return
	}

	for _, r := range match[len(file):] {
		if doesRuneBreakIdentifier(r) {
			c.completion = append(c.completion, '\\')
		}

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

func getPathExecutables() []string {
	var executables []string
	seen := make(map[string]bool)

	pathEnv := os.Getenv("PATH")
	dirs := filepath.SplitList(pathEnv)

	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)

		if err != nil {
			continue
		}

		for _, entry := range entries {

			if entry.IsDir() {
				continue
			}

			info, err := entry.Info()

			if err != nil {
				continue
			}

			if info.Mode()&0111 != 0 {
				name := entry.Name()

				if !seen[name] {
					executables = append(executables, name)
					seen[name] = true
				}
			}
		}
	}

	return executables
}

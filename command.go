package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type command struct {
	runes []rune

	tokenizer tokenizer
	parser    parser
	isDirty   bool

	completion      []rune
	pathExecutables []string

	history          [][]rune
	historyIndex     int
	pendingRunes     []rune
	historicalUsages map[string]int
}

func newCommand() command {
	pathExecutables := getPathExecutables()
	historicalUsages := make(map[string]int)

	return command{
		pathExecutables:  pathExecutables,
		historicalUsages: historicalUsages,
	}
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

func (c *command) addToHistory() {
	if len(c.runes) == 0 {
		return
	}

	ast, didSucceed := c.parse()

	if !didSucceed {
		panic("Only successful commands should be added to the history")
	}

	ast.analyze(c)

	if len(c.history) == 0 || !slices.Equal(c.history[len(c.history)-1], c.runes) {
		item := make([]rune, len(c.runes))
		copy(item, c.runes)

		c.history = append(c.history, item)

	}

	c.historyIndex = len(c.history)
	c.pendingRunes = nil
}

func (c *command) historyUp() {
	if c.historyIndex == 0 {
		return
	}

	if c.historyIndex == len(c.history) {
		c.pendingRunes = make([]rune, len(c.runes))
		copy(c.pendingRunes, c.runes)
	}

	c.historyIndex--
	c.loadHistory(c.historyIndex)
}

func (c *command) historyDown() {
	if c.historyIndex >= len(c.history) {
		return
	}

	c.historyIndex++

	if c.historyIndex == len(c.history) {
		c.runes = make([]rune, len(c.pendingRunes))
		copy(c.runes, c.pendingRunes)

		c.isDirty = true
	} else {
		c.loadHistory(c.historyIndex)
	}
}

func (c *command) loadHistory(index int) {
	item := c.history[index]

	c.runes = make([]rune, len(item))
	copy(c.runes, item)

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
	match, isDir := completeFilePath(path)

	if len(call.children) == 1 && c.runes[len(c.runes)-1] != ' ' {
		prefix := call.children[0]
		executableMatch := c.completeExecutable(prefix)

		if isBetterMatch(match, executableMatch) {
			match, isDir = executableMatch, false
		}
	}

	for _, r := range match {
		if doesRuneBreakIdentifier(r) {
			c.completion = append(c.completion, '\\')
		}

		c.completion = append(c.completion, r)
	}

	if isDir {
		c.completion = append(c.completion, '/')
	}
}

func completeFilePath(path string) (string, bool) {
	dir, file := filepath.Split(path)

	if dir == "" {
		dir = "."
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		return "", false
	}

	match := ""
	isDir := false

	for _, entry := range entries {
		name := entry.Name()

		if isBetterMatch(match, name) && strings.HasPrefix(name, file) {
			match = name
			isDir = entry.IsDir()
		}
	}

	if len(match) < len(file) {
		return "", false
	}

	return match[len(file):], isDir
}

func (c *command) completeExecutable(prefix string) string {
	match := ""

	for _, executable := range c.pathExecutables {
		if !strings.HasPrefix(executable, prefix) {
			continue
		}

		if isBetterMatch(match, executable) {
			match = executable
			continue
		}

		if len(match) == len(executable) && c.historicalUsages[match] < c.historicalUsages[executable] {
			match = executable
		}
	}

	if len(match) < len(prefix) {
		return ""
	}

	return match[len(prefix):]
}

func isBetterMatch(oldMatch, newMatch string) bool {
	return newMatch != "" && (oldMatch == "" || len(newMatch) < len(oldMatch))
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

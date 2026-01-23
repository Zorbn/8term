package main

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
)

type command struct {
	runes []rune

	cursorIndex int

	tokenizer            tokenizer
	parser               parser
	missingTrailingRunes []rune
	isDirty              bool

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
		cursorIndex:      0,
		pathExecutables:  pathExecutables,
		historicalUsages: historicalUsages,
	}
}

func (c *command) insert(r rune) {
	c.runes = slices.Insert(c.runes, c.cursorIndex, r)
	c.cursorIndex++
	c.isDirty = true
}

func (c *command) pop() {
	if len(c.runes) == 0 || c.cursorIndex == 0 {
		return
	}

	c.runes = slices.Delete(c.runes, c.cursorIndex-1, c.cursorIndex)
	c.cursorIndex--
	c.isDirty = true
}

func (c *command) clear() {
	if len(c.runes) == 0 {
		return
	}

	c.runes = c.runes[:0]
	c.cursorIndex = 0
	c.isDirty = true
}

func (c *command) moveCursorLeft() {
	if c.cursorIndex <= 0 {
		return
	}

	c.cursorIndex--
	c.updateCompletion()
}

func (c *command) moveCursorRight() {
	if c.cursorIndex >= len(c.runes) {
		return
	}

	c.cursorIndex++
	c.updateCompletion()
}

func (c *command) getCursorRune() rune {
	if len(c.completion) > 0 {
		return c.completion[0]
	}

	if c.cursorIndex < len(c.runes) {
		return c.runes[c.cursorIndex]
	}

	if len(c.missingTrailingRunes) > 0 {
		return c.missingTrailingRunes[0]
	}

	return ' '
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
	c.cursorIndex = 0
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
		c.cursorIndex = len(c.runes)

		c.isDirty = true
	} else {
		c.loadHistory(c.historyIndex)
	}
}

func (c *command) loadHistory(index int) {
	item := c.history[index]

	c.runes = make([]rune, len(item))
	copy(c.runes, item)
	c.cursorIndex = len(c.runes)

	c.isDirty = true
}

func (c *command) parse() (astNode, bool) {
	if c.isDirty {
		c.isDirty = false

		c.missingTrailingRunes = c.missingTrailingRunes[:0]
		c.tokenizer.tokenize(c.runes, c.missingTrailingRunes)
		c.parser.parse(c.tokenizer.tokens, c.missingTrailingRunes)
		c.updateCompletion()
	}

	return c.parser.ast, c.tokenizer.didSucceed && len(c.parser.errors) == 0
}

func (c *command) updateCompletion() {
	ast, _ := c.parse()
	c.completion = c.completion[:0]

	var token *token

	for i := range c.tokenizer.tokens {
		t := &c.tokenizer.tokens[i]

		if c.cursorIndex == t.end {
			token = t
			break
		}
	}

	if token == nil {
		return
	}

	node := ast.find(c.cursorIndex)
	needsExecutable := false

	switch n := node.(type) {
	case *callNode:
		needsExecutable = len(n.children) == 0 || n.children[0] == *token
	}

	match, isDir := completeFilePath(token.text, needsExecutable)

	if needsExecutable {
		executableMatch := c.completeExecutable(token.text)

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

func completeFilePath(path string, needsExecutable bool) (string, bool) {
	dir, file := filepath.Split(path)

	if dir == "" {
		if needsExecutable {
			return "", false
		}

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

		if needsExecutable && !isEntryExecutable(entry) {
			continue
		}

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
		c.insert(r)
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
			if !isEntryExecutable(entry) {
				continue
			}

			name := entry.Name()

			if !seen[name] {
				executables = append(executables, name)
				seen[name] = true
			}
		}
	}

	return executables
}

func isEntryExecutable(entry os.DirEntry) bool {
	if entry.IsDir() {
		return false
	}

	info, err := entry.Info()

	if err != nil {
		return false
	}

	return info.Mode()&0111 != 0
}

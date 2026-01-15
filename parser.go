package main

import "errors"

type astNode interface {
	exec(pane *pane) int
}

type callNode struct {
	tokens []string
}

type binaryNode struct {
	op    string
	left  astNode
	right astNode
}

type pipeNode struct {
	children []*callNode
}

type parser struct {
	tokens               []token
	pos                  int
	missingTrailingRunes []rune
	errors               []error
}

func (p *parser) parse(tokens []token) astNode {
	p.tokens = tokens
	p.pos = 0
	p.missingTrailingRunes = p.missingTrailingRunes[:0]
	p.errors = p.errors[:0]

	node := p.parseSequence()

	if p.pos < len(p.tokens) {
		p.error("Unexpected token: " + p.peek())
	}

	return node
}

func (p *parser) error(text string) {
	p.errors = append(p.errors, errors.New(text))
}

func (p *parser) peek() string {
	if p.pos >= len(p.tokens) {
		return ""
	}

	return string(p.tokens[p.pos])
}

func (p *parser) consume() string {
	s := p.peek()

	if p.pos < len(p.tokens) {
		p.pos++
	}

	return s
}

func (p *parser) match(s string) bool {
	if p.peek() == s {
		p.consume()
		return true
	}

	return false
}

func (p *parser) parseSequence() astNode {
	left := p.parseLogic()

	for p.match(";") {

		if p.peek() == "" {
			break
		}

		right := p.parseLogic()
		left = &binaryNode{op: ";", left: left, right: right}
	}

	return left
}

func (p *parser) parseLogic() astNode {
	left := p.parseTerm()

	for {
		op := p.peek()

		if op == "&&" || op == "||" {
			p.consume()
			right := p.parseTerm()

			left = &binaryNode{op: op, left: left, right: right}
		} else {
			break
		}
	}

	return left
}

func (p *parser) parseTerm() astNode {
	if p.match("(") {
		node := p.parseSequence()

		if !p.match(")") {
			p.missingTrailingRunes = append(p.missingTrailingRunes, ')')
		}

		return node
	}

	return p.parsePipe()
}

func (p *parser) parsePipe() astNode {
	first := p.parseCall()

	if p.peek() != "|" {
		return first
	}

	children := []*callNode{first}

	for p.match("|") {
		next := p.parseCall()
		children = append(children, next)
	}

	return &pipeNode{children: children}
}

func (p *parser) parseCall() *callNode {
	if p.peek() == "" {
		p.error("Unexpected end of input")
		return &callNode{}
	}

	var tokens []string

	for {
		token := p.peek()

		if token == "" || isOperator(token) {
			break
		}

		if isRuneSpecial(rune(token[0])) {
			p.error("Expected string but got symbol")
		}

		tokens = append(tokens, p.consume())
	}

	if len(tokens) == 0 {
		p.error("Empty call")
	}

	return &callNode{tokens}
}

func isOperator(s string) bool {
	switch s {
	case ";", "&&", "||", "|":
		return true
	}

	return false
}

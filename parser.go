package main

import (
	"errors"
)

type astNode interface {
	exec(pane *pane) int
	analyze(command *command)
	find(index int) astNode
	Start() int
	End() int
}

type callNode struct {
	children []token
	start    int
	end      int
}

func (c *callNode) Start() int { return c.start }
func (c *callNode) End() int   { return c.end }

type binaryNode struct {
	op    string
	left  astNode
	right astNode
	start int
	end   int
}

func (b *binaryNode) Start() int { return b.start }
func (b *binaryNode) End() int   { return b.end }

type pipeNode struct {
	children []*callNode
	start    int
	end      int
}

func (p *pipeNode) Start() int { return p.start }
func (p *pipeNode) End() int   { return p.end }

type parser struct {
	pos          int
	errors       []error
	ast          astNode
	lastCallNode *callNode
}

func (p *parser) parse(tokens []token, missingTrailingRunes []rune) {
	p.pos = 0
	p.errors = p.errors[:0]
	p.lastCallNode = nil

	ast := p.parseSequence(tokens, missingTrailingRunes)

	if p.pos < len(tokens) {
		p.error("Unexpected token: " + string(p.peek(tokens).text))
	}

	p.ast = ast
}

func (p *parser) error(text string) {
	p.errors = append(p.errors, errors.New(text))
}

func (p *parser) peek(tokens []token) token {
	if p.pos >= len(tokens) {
		return newToken("", tokenKindEof, 0, 0)
	}

	return tokens[p.pos]
}

func (p *parser) consume(tokens []token) token {
	token := p.peek(tokens)

	if p.pos < len(tokens) {
		p.pos++
	}

	return token
}

func (p *parser) match(text string, kind tokenKind, tokens []token) bool {
	t := p.peek(tokens)
	if t.text == text && t.kind == kind {
		p.consume(tokens)
		return true
	}

	return false
}

func (p *parser) parseSequence(tokens []token, missingTrailingRunes []rune) astNode {
	left := p.parseLogic(tokens, missingTrailingRunes)

	for p.match(";", tokenKindSymbol, tokens) {

		if p.peek(tokens).kind == tokenKindEof {
			break
		}

		right := p.parseLogic(tokens, missingTrailingRunes)
		left = &binaryNode{op: ";", left: left, right: right, start: left.Start(), end: right.End()}
	}

	return left
}

func (p *parser) parseLogic(tokens []token, missingTrailingRunes []rune) astNode {
	left := p.parseTerm(tokens, missingTrailingRunes)

	for {
		op := p.peek(tokens)

		if op.isSymbol("&&") || op.isSymbol("||") {
			p.consume(tokens)
			right := p.parseTerm(tokens, missingTrailingRunes)

			left = &binaryNode{op: op.text, left: left, right: right, start: left.Start(), end: right.End()}
		} else {
			break
		}
	}

	return left
}

func (p *parser) parseTerm(tokens []token, missingTrailingRunes []rune) astNode {
	if p.match("(", tokenKindSymbol, tokens) {
		node := p.parseSequence(tokens, missingTrailingRunes)

		if !p.match(")", tokenKindSymbol, tokens) {
			missingTrailingRunes = append(missingTrailingRunes, ')')
		}

		return node
	}

	return p.parsePipe(tokens)
}

func (p *parser) parsePipe(tokens []token) astNode {
	first := p.parseCall(tokens)

	if !p.peek(tokens).isSymbol("|") {
		return first
	}

	children := []*callNode{first}

	for p.match("|", tokenKindSymbol, tokens) {
		next := p.parseCall(tokens)
		children = append(children, next)
	}

	return &pipeNode{children: children, start: first.Start(), end: children[len(children)-1].End()}
}

func (p *parser) parseCall(tokens []token) *callNode {
	callNode := p.parseCallInner(tokens)
	p.lastCallNode = callNode

	return callNode
}

func (p *parser) parseCallInner(tokens []token) *callNode {
	if p.peek(tokens).kind == tokenKindEof {
		p.error("Unexpected end of input")
		return &callNode{}
	}

	var children []token

	for {
		token := p.peek(tokens)

		if token.kind == tokenKindEof || isOperator(token) {
			break
		}

		if token.kind != tokenKindString {
			p.error("Expected string")
		}

		t := p.consume(tokens)
		children = append(children, t)
	}

	start := 0
	end := 0

	if len(children) == 0 {
		p.error("Empty call")
	} else {
		start = children[0].start
		end = children[len(children)-1].end
	}

	return &callNode{children: children, start: start, end: end}
}

func isOperator(token token) bool {
	if token.kind != tokenKindSymbol {
		return false
	}

	switch token.text {
	case ";", "&&", "||", "|":
		return true
	}

	return false
}

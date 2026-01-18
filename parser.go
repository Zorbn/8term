package main

import (
	"errors"
)

type astNode interface {
	exec(pane *pane) int
	analyze(command *command)
}

type callNode struct {
	children []string
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
	pos                  int
	missingTrailingRunes []rune
	errors               []error
	ast                  astNode
	lastCallNode         *callNode
}

func (p *parser) parse(tokens []token) {
	p.pos = 0
	p.missingTrailingRunes = p.missingTrailingRunes[:0]
	p.errors = p.errors[:0]
	p.lastCallNode = nil

	ast := p.parseSequence(tokens)

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
		return newToken("", tokenKindEof)
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
	if p.peek(tokens) == newToken(text, kind) {
		p.consume(tokens)
		return true
	}

	return false
}

func (p *parser) parseSequence(tokens []token) astNode {
	left := p.parseLogic(tokens)

	for p.match(";", tokenKindSymbol, tokens) {

		if p.peek(tokens).kind == tokenKindEof {
			break
		}

		right := p.parseLogic(tokens)
		left = &binaryNode{op: ";", left: left, right: right}
	}

	return left
}

func (p *parser) parseLogic(tokens []token) astNode {
	left := p.parseTerm(tokens)

	for {
		op := p.peek(tokens)

		if op.isSymbol("&&") || op.isSymbol("||") {
			p.consume(tokens)
			right := p.parseTerm(tokens)

			left = &binaryNode{op: op.text, left: left, right: right}
		} else {
			break
		}
	}

	return left
}

func (p *parser) parseTerm(tokens []token) astNode {
	if p.match("(", tokenKindSymbol, tokens) {
		node := p.parseSequence(tokens)

		if !p.match(")", tokenKindSymbol, tokens) {
			p.missingTrailingRunes = append(p.missingTrailingRunes, ')')
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

	return &pipeNode{children: children}
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

	var children []string

	for {
		token := p.peek(tokens)

		if token.kind == tokenKindEof || isOperator(token) {
			break
		}

		if token.kind != tokenKindString {
			p.error("Expected string")
		}

		children = append(children, p.consume(tokens).text)
	}

	if len(tokens) == 0 {
		p.error("Empty call")
	}

	return &callNode{children}
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

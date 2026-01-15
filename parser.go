package main

import (
	"errors"
)

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
	tokens []token
	pos    int
}

func parse(tokens []token) (astNode, error) {
	p := &parser{tokens: tokens}
	node, err := p.parseSequence()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		return nil, errors.New("unexpected token: " + p.peek())
	}
	return node, nil
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

func (p *parser) parseSequence() (astNode, error) {
	left, err := p.parseLogic()
	if err != nil {
		return nil, err
	}

	for p.match(";") {
		if p.peek() == "" {
			break
		}
		right, err := p.parseLogic()
		if err != nil {
			return nil, err
		}
		left = &binaryNode{op: ";", left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseLogic() (astNode, error) {
	left, err := p.parsePipe()
	if err != nil {
		return nil, err
	}

	for {
		op := p.peek()
		if op == "&&" || op == "||" {
			p.consume()
			right, err := p.parsePipe()
			if err != nil {
				return nil, err
			}
			left = &binaryNode{op: op, left: left, right: right}
		} else {
			break
		}
	}
	return left, nil
}

func (p *parser) parsePipe() (astNode, error) {
	first, err := p.parseCall()
	if err != nil {
		return nil, err
	}

	if p.peek() != "|" {
		return first, nil
	}

	children := []*callNode{first}
	for p.match("|") {
		next, err := p.parseCall()
		if err != nil {
			return nil, err
		}
		children = append(children, next)
	}
	return &pipeNode{children: children}, nil
}

func (p *parser) parseTerm() (astNode, error) {
	if p.match("(") {
		node, err := p.parseSequence()
		if err != nil {
			return nil, err
		}
		if !p.match(")") {
			return nil, errors.New("expected ')'")
		}
		return node, nil
	}
	return p.parseCall()
}

func (p *parser) parseCall() (*callNode, error) {
	if isOperator(p.peek()) {
		return nil, errors.New("expected command, found " + p.peek())
	}
	if p.peek() == "" {
		return nil, errors.New("unexpected end of input")
	}

	var args []string
	for {
		tok := p.peek()
		if tok == "" || isOperator(tok) || tok == ")" {
			break
		}
		args = append(args, p.consume())
	}

	if len(args) == 0 {
		return nil, errors.New("empty call")
	}

	return &callNode{tokens: args}, nil
}

func isOperator(s string) bool {
	switch s {
	case ";", "&&", "||", "|", "(", ")":
		return true
	}
	return false
}

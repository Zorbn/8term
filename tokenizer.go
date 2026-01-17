package main

import (
	"strings"
	"unicode"
)

type tokenKind int

const (
	tokenKindString tokenKind = iota
	tokenKindSymbol
	tokenKindEof
)

type token struct {
	text string
	kind tokenKind
}

func newToken(text string, kind tokenKind) token {
	return token{text, kind}
}

func (t token) isSymbol(text string) bool {
	return t.kind == tokenKindSymbol && t.text == text
}

type tokenizer struct {
	tokens               []token
	missingTrailingRunes []rune
	didSucceed           bool
	tokenText            strings.Builder
}

func (t *tokenizer) tokenize(text []rune) {
	t.tokens = t.tokens[:0]
	t.missingTrailingRunes = t.missingTrailingRunes[:0]
	t.didSucceed = true

	var token token
	r := ' '

	for len(text) > 0 {
		r, text = nextRune(text)

		if unicode.IsSpace(r) {
			continue
		}

		switch r {
		case '"':
			token, text = t.tokenizeString(text, '"', '\\')
		case '\'':
			token, text = t.tokenizeString(text, '\'', 0)
		default:
			if isRuneSpecial(r) {
				didSucceed := false
				token, text, didSucceed = t.tokenizeSymbol(text, r)

				if !didSucceed {
					t.didSucceed = false
				}
			} else {
				token, text = t.tokenizeIdentifier(text, r)
			}
		}

		t.tokens = append(t.tokens, token)
	}
}

func (t *tokenizer) newToken(kind tokenKind) token {
	token := newToken(t.tokenText.String(), kind)
	t.tokenText.Reset()

	return token
}

func (t *tokenizer) tokenizeString(text []rune, delimiter rune, escape rune) (token, []rune) {
	r := delimiter
	isEscaped := false

	for len(text) > 0 {
		r, text = nextRune(text)

		if r == escape {
			isEscaped = true
			continue
		}

		if isEscaped {
			if r == delimiter {
				continue
			}

			t.tokenText.WriteRune(escape)
		}

		isEscaped = false

		if r == delimiter {
			return t.newToken(tokenKindString), text
		}

		t.tokenText.WriteRune(r)
	}

	t.missingTrailingRunes = append(t.missingTrailingRunes, delimiter)

	return t.newToken(tokenKindString), text
}

func (t *tokenizer) tokenizeIdentifier(text []rune, firstRune rune) (token, []rune) {
	t.tokenText.WriteRune(firstRune)

	r := firstRune

	for len(text) > 0 {
		if unicode.IsSpace(text[0]) || isRuneSpecial(text[0]) {
			return t.newToken(tokenKindString), text
		}

		r, text = nextRune(text)

		switch r {
		case '"', '\'':
			return t.newToken(tokenKindString), text
		}

		t.tokenText.WriteRune(r)
	}

	return t.newToken(tokenKindString), text
}

func (t *tokenizer) tokenizeSymbol(text []rune, firstRune rune) (token, []rune, bool) {
	t.tokenText.WriteRune(firstRune)

	var secondRune rune

	switch firstRune {
	case ';', '(', ')':
		return t.newToken(tokenKindSymbol), text, true
	case '|':
		if len(text) > 0 && text[0] == '|' {
			secondRune, text = nextRune(text)
			t.tokenText.WriteRune(secondRune)

			return t.newToken(tokenKindSymbol), text, true
		}

		return t.newToken(tokenKindSymbol), text, true
	case '&':
		if len(text) > 0 && text[0] == '&' {
			secondRune, text = nextRune(text)
			t.tokenText.WriteRune(secondRune)

			return t.newToken(tokenKindSymbol), text, true
		}

		return t.newToken(tokenKindSymbol), text, false
	default:
		panic("Unexpected rune in symbol")
	}
}

func nextRune(text []rune) (rune, []rune) {
	return text[0], text[1:]
}

func isRuneSpecial(r rune) bool {
	switch r {
	case '"', '\'', ';', '|', '&', '(', ')':
		return true
	}

	return false
}

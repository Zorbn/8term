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
	text  string
	kind  tokenKind
	start int
	end   int
}

func newToken(text string, kind tokenKind, start int, end int) token {
	return token{text, kind, start, end}
}

func (t token) isSymbol(text string) bool {
	return t.kind == tokenKindSymbol && t.text == text
}

type tokenizer struct {
	tokens     []token
	didSucceed bool
	tokenText  strings.Builder
	cursor     int
}

func (t *tokenizer) tokenize(text []rune, missingTrailingRunes []rune) {
	t.tokens = t.tokens[:0]
	t.didSucceed = true
	t.cursor = 0

	var token token
	r := ' '

	for len(text) > 0 {
		r, text = t.nextRune(text)

		if unicode.IsSpace(r) {
			continue
		}

		start := t.cursor - 1

		switch r {
		case '"':
			token, text = t.tokenizeString(text, '"', true, start, missingTrailingRunes)
		case '\'':
			token, text = t.tokenizeString(text, '\'', false, start, missingTrailingRunes)
		default:
			if isRuneSpecial(r) {
				didSucceed := false
				token, text, didSucceed = t.tokenizeSymbol(text, r, start)

				if !didSucceed {
					t.didSucceed = false
				}
			} else {
				token, text = t.tokenizeIdentifier(text, r, start)
			}
		}

		t.tokens = append(t.tokens, token)
	}
}

func (t *tokenizer) newToken(kind tokenKind, start int) token {
	token := newToken(t.tokenText.String(), kind, start, t.cursor)
	t.tokenText.Reset()

	return token
}

func (t *tokenizer) tokenizeString(text []rune, delimiter rune, canEscape bool, start int, missingTrailingRunes []rune) (token, []rune) {
	r := delimiter
	isEscaped := false

	for len(text) > 0 {
		r, text = t.nextRune(text)

		if canEscape && r == '\\' {
			isEscaped = true
			continue
		}

		if isEscaped {
			if r == delimiter {
				continue
			}

			t.tokenText.WriteRune('\\')
		}

		isEscaped = false

		if r == delimiter {
			return t.newToken(tokenKindString, start), text
		}

		t.tokenText.WriteRune(r)
	}

	missingTrailingRunes = append(missingTrailingRunes, delimiter)

	return t.newToken(tokenKindString, start), text
}

func (t *tokenizer) tokenizeIdentifier(text []rune, firstRune rune, start int) (token, []rune) {
	t.tokenText.WriteRune(firstRune)

	r := firstRune
	isEscaped := false

	for len(text) > 0 {
		if !isEscaped && doesRuneBreakIdentifier(text[0]) {
			return t.newToken(tokenKindString, start), text
		}

		isEscaped = false

		r, text = t.nextRune(text)

		switch r {
		case '\\':
			isEscaped = true
			continue
		case '"', '\'':
			return t.newToken(tokenKindString, start), text
		}

		t.tokenText.WriteRune(r)
	}

	return t.newToken(tokenKindString, start), text
}

func (t *tokenizer) tokenizeSymbol(text []rune, firstRune rune, start int) (token, []rune, bool) {
	t.tokenText.WriteRune(firstRune)

	var secondRune rune

	switch firstRune {
	case ';', '(', ')':
		return t.newToken(tokenKindSymbol, start), text, true
	case '|':
		if len(text) > 0 && text[0] == '|' {
			secondRune, text = t.nextRune(text)
			t.tokenText.WriteRune(secondRune)

			return t.newToken(tokenKindSymbol, start), text, true
		}

		return t.newToken(tokenKindSymbol, start), text, true
	case '&':
		if len(text) > 0 && text[0] == '&' {
			secondRune, text = t.nextRune(text)
			t.tokenText.WriteRune(secondRune)

			return t.newToken(tokenKindSymbol, start), text, true
		}

		return t.newToken(tokenKindSymbol, start), text, false
	default:
		panic("Unexpected rune in symbol")
	}
}

func (t *tokenizer) nextRune(text []rune) (rune, []rune) {
	t.cursor++
	return text[0], text[1:]
}

func doesRuneBreakIdentifier(r rune) bool {
	return unicode.IsSpace(r) || isRuneSpecial(r)
}

func isRuneSpecial(r rune) bool {
	switch r {
	case '"', '\'', ';', '|', '&', '(', ')':
		return true
	}

	return false
}

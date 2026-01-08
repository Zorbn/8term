package main

import "unicode"

type token []rune

func tokenize(text []rune) []token {
	var tokens []token
	var t token
	r := ' '

	for len(text) > 0 {
		r, text = nextRune(text)

		if unicode.IsSpace(r) {
			continue
		}

		switch r {
		case '"':
			t, text = tokenizeString(text, '"', '\\')
		case '\'':
			t, text = tokenizeString(text, '\'', 0)
		default:
			t, text = tokenizeIdentifier(text, r)
		}

		tokens = append(tokens, t)
	}

	return tokens
}

func tokenizeString(text []rune, delimiter rune, escape rune) (token, []rune) {
	t := make(token, 0)
	r := delimiter
	isEscaped := false

	for len(text) > 0 {
		r, text = nextRune(text)

		if !isEscaped && r == delimiter {
			return t, text
		}

		isEscaped = r == escape

		if isEscaped {
			continue
		}

		t = append(t, r)
	}

	// TODO: The string wasn't terminated, this should be an error! Probably return nil.
	return t, text
}

func tokenizeIdentifier(text []rune, firstRune rune) (token, []rune) {
	t := token{firstRune}
	r := firstRune

	for len(text) > 0 {
		r, text = nextRune(text)

		if unicode.IsSpace(r) || r == '"' || r == '\'' {
			return t, text
		}

		switch r {
		case '"', '\'':
			return t, text
		}

		t = append(t, r)
	}

	return t, text
}

func nextRune(text []rune) (rune, []rune) {
	return text[0], text[1:]
}

package main

import "unicode"

type token []rune

type tokenizeResult struct {
	tokens               []token
	missingTrailingRunes []rune
}

func tokenize(text []rune, result *tokenizeResult) {
	result.tokens = result.tokens[:0]
	result.missingTrailingRunes = result.missingTrailingRunes[:0]

	var t token
	r := ' '

	for len(text) > 0 {
		r, text = nextRune(text)

		if unicode.IsSpace(r) {
			continue
		}

		hadDelimiter := false

		switch r {
		case '"':
			t, text, hadDelimiter = tokenizeString(text, '"', '\\')

			if !hadDelimiter {
				result.missingTrailingRunes = append(result.missingTrailingRunes, '"')
			}
		case '\'':
			t, text, hadDelimiter = tokenizeString(text, '\'', 0)

			if !hadDelimiter {
				result.missingTrailingRunes = append(result.missingTrailingRunes, '\'')
			}
		default:
			t, text = tokenizeIdentifier(text, r)
		}

		result.tokens = append(result.tokens, t)
	}
}

func tokenizeString(text []rune, delimiter rune, escape rune) (token, []rune, bool) {
	t := make(token, 0)
	r := delimiter
	isEscaped := false

	for len(text) > 0 {
		r, text = nextRune(text)

		if !isEscaped && r == delimiter {
			return t, text, true
		}

		isEscaped = r == escape

		if isEscaped {
			continue
		}

		t = append(t, r)
	}

	return t, text, false
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

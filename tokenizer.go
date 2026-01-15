package main

import "unicode"

type token []rune

// TODO: Make this the tokenizer that has methods, like how parser works.
// TODO: We're generally converting tokens to strings, should tokens be strings to start with?
// TODO: There's no way to distinguish between && (operator) and "&&" (string).
type tokenizeResult struct {
	tokens               []token
	missingTrailingRunes []rune
	didSucceed           bool
}

func tokenize(text []rune, result *tokenizeResult) {
	result.tokens = result.tokens[:0]
	result.missingTrailingRunes = result.missingTrailingRunes[:0]
	result.didSucceed = true

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
			if isRuneSpecial(r) {
				didSucceed := false
				t, text, didSucceed = tokenizeSymbol(text, r)

				if !didSucceed {
					result.didSucceed = false
				}
			} else {
				t, text = tokenizeIdentifier(text, r)
			}
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

		if r == escape {
			isEscaped = true
			continue
		}

		if isEscaped {
			if r == delimiter {
				continue
			}

			t = append(t, escape)
		}

		isEscaped = false

		if r == delimiter {
			return t, text, true
		}

		t = append(t, r)
	}

	return t, text, false
}

func tokenizeIdentifier(text []rune, firstRune rune) (token, []rune) {
	t := token{firstRune}
	r := firstRune

	for len(text) > 0 {
		if unicode.IsSpace(text[0]) || isRuneSpecial(text[0]) {
			return t, text
		}

		r, text = nextRune(text)

		switch r {
		case '"', '\'':
			return t, text
		}

		t = append(t, r)
	}

	return t, text
}

func tokenizeSymbol(text []rune, firstRune rune) (token, []rune, bool) {
	var secondRune rune

	switch firstRune {
	case ';', '(', ')':
		return token{firstRune}, text, true
	case '|':
		if len(text) > 0 && text[0] == '|' {
			secondRune, text = nextRune(text)
			return token{firstRune, secondRune}, text, true
		}

		return token{firstRune}, text, false
	case '&':
		if len(text) > 0 && text[0] == '&' {
			secondRune, text = nextRune(text)
			return token{firstRune, secondRune}, text, true
		}

		return token{firstRune}, text, false
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

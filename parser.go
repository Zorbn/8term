package main

type Performer interface {
	Plain(text []byte)
	Backspace()
	Tab()
	CarriageReturn()
	Newline()
	ReverseNewline()
	HideCursor()
	ShowCursor()
	SwitchToNormalBuffer()
	SwitchToAlternateBuffer()
	QueryModifyKeyboard()
	QueryModifyCursorKeys()
	QueryModifyFunctionKeys()
	QueryModifyOtherKeys()
	QueryDeviceAttributes()
	ResetFormatting()
	SetColorsBright(bright bool)
	SetColorsSwapped(swapped bool)
	SetForegroundColor(color uint32)
	SetBackgroundColor(color uint32)
	SetCursorX(x int)
	SetCursorY(y int)
	SetCursorPosition(x, y int)
	MoveCursorX(delta int)
	MoveCursorY(delta int)
	MoveCursorYAndResetX(delta int)
	ClearToScreenEnd()
	ClearToScreenStart()
	ClearScreen()
	ClearScrollbackLines()
	ClearToLineEnd()
	ClearToLineStart()
	ClearLine()
	InsertLines(count int)
	DeleteLines(count int)
	ScrollUp(count int)
	ScrollDown(count int)
	ClearCharsAfterCursor(count int)
	DeleteCharsAfterCursor(count int)
	SetScrollRegion(top, bottom int)
	QueryDeviceStatus()
	QueryTerminalId()
	SetTitle(title []byte)
	ResetTitle()
	QueryForegroundColor()
	QueryBackgroundColor()
}

func ParseEscapeSequences(output []byte, performer Performer) {
	plainOutput := output

	for len(output) > 0 {
		resetPlainOutput := true

		switch output[0] {
		case 0x1B: // ESC
			flushPlainOutput(plainOutput, output, performer)

			if len(output) < 2 {
				output = output[1:]
				continue
			}

			var remaining []byte
			switch output[1] {
			case '[':
				remaining = parseEscapeSequencesCSI(output[2:], performer)
			case ']':
				remaining = parseEscapeSequencesOSC(output[2:], performer)
			case '(':
				if len(output) > 2 && output[2] == 'B' {
					// Use ASCII character set (other character sets are unsupported)
					remaining = output[3:]
				}
			case '=':
				// Enter alternative keypad mode, ignored
				remaining = output[2:]
			case '>':
				// Exit alternative keypad mode, ignored
				remaining = output[2:]
			case 'M':
				performer.ReverseNewline()
				remaining = output[2:]
			}

			if remaining != nil {
				output = remaining
			} else {
				output = output[1:]
			}

		case 0x7: // Bell
			flushPlainOutput(plainOutput, output, performer)
			output = output[1:]

		case 0x8: // Backspace
			flushPlainOutput(plainOutput, output, performer)
			performer.Backspace()
			output = output[1:]

		case '\t': // Tab
			flushPlainOutput(plainOutput, output, performer)
			performer.Tab()
			output = output[1:]

		case '\r': // Carriage Return
			flushPlainOutput(plainOutput, output, performer)
			performer.CarriageReturn()
			output = output[1:]

		case '\n': // Newline
			flushPlainOutput(plainOutput, output, performer)
			performer.Newline()
			output = output[1:]

		default:
			output = output[1:]
			resetPlainOutput = false
		}

		if resetPlainOutput {
			plainOutput = output
		}
	}

	flushPlainOutput(plainOutput, output, performer)
}

func flushPlainOutput(plainOutput, output []byte, performer Performer) {
	plainLen := len(plainOutput) - len(output)

	if plainLen == 0 {
		return
	}

	performer.Plain(plainOutput[:plainLen])
}

func parseEscapeSequencesCSI(output []byte, performer Performer) []byte {
	if len(output) == 0 {
		return nil
	}

	switch output[0] {
	case '?':
		return parsePrefixedCSI(output[1:], performer, '?')
	case '>':
		return parsePrefixedCSI(output[1:], performer, '>')
	default:
		return parseUnprefixedCSI(output, performer)
	}
}

func parsePrefixedCSI(output []byte, performer Performer, prefix byte) []byte {
	var params []int
	params, output = parseNumericParameters(output)

	if len(output) == 0 {
		return nil
	}

	cmd := output[0]
	output = output[1:]

	if prefix == '?' {
		switch cmd {
		case 'l':
			if len(params) > 0 {
				switch params[0] {
				case 25:
					performer.HideCursor()
				case 1047, 1049:
					performer.SwitchToNormalBuffer()
				}
			}
			return output
		case 'h':
			if len(params) > 0 {
				switch params[0] {
				case 25:
					performer.ShowCursor()
				case 1047, 1049:
					performer.SwitchToAlternateBuffer()
				}
			}
			return output
		case 'm':
			if len(params) > 0 {
				switch params[0] {
				case 0:
					performer.QueryModifyKeyboard()
				case 1:
					performer.QueryModifyCursorKeys()
				case 2:
					performer.QueryModifyFunctionKeys()
				case 4:
					performer.QueryModifyOtherKeys()
				default:
					return nil
				}
			}
			return output
		}
	} else if prefix == '>' {
		switch cmd {
		case 'm':
			// Set/reset xterm modifier key options, ignored
			return output
		case 'c':
			performer.QueryDeviceAttributes()
			return output
		}
	}

	return nil
}

func parseUnprefixedCSI(output []byte, performer Performer) []byte {
	var params []int
	params, output = parseNumericParameters(output)
	if len(output) == 0 {
		return nil
	}

	cmd := output[0]
	output = output[1:]

	switch cmd {
	case 'm':
		parseFormatting(params, performer)
		return output

	case 'l', 'h':
		// Mode disabled/enabled, ignored
		return output

	case 'G':
		x := parameter(params, 0, 1) - 1
		if x < 0 {
			x = 0
		}
		performer.SetCursorX(x)
		return output

	case 'd':
		y := parameter(params, 0, 1) - 1
		if y < 0 {
			y = 0
		}
		performer.SetCursorY(y)
		return output

	case 'H':
		y := parameter(params, 0, 1) - 1
		x := parameter(params, 1, 1) - 1
		if y < 0 {
			y = 0
		}
		if x < 0 {
			x = 0
		}
		performer.SetCursorPosition(x, y)
		return output

	case 'A':
		distance := parameter(params, 0, 1)
		performer.MoveCursorY(-distance)
		return output

	case 'B':
		distance := parameter(params, 0, 1)
		performer.MoveCursorY(distance)
		return output

	case 'C':
		distance := parameter(params, 0, 1)
		performer.MoveCursorX(distance)
		return output

	case 'D':
		distance := parameter(params, 0, 1)
		performer.MoveCursorX(-distance)
		return output

	case 'E':
		distance := parameter(params, 0, 1)
		performer.MoveCursorYAndResetX(distance)
		return output

	case 'F':
		distance := parameter(params, 0, 1)
		performer.MoveCursorYAndResetX(-distance)
		return output

	case 'J':
		mode := parameter(params, 0, 0)
		switch mode {
		case 0:
			performer.ClearToScreenEnd()
		case 1:
			performer.ClearToScreenStart()
		case 2:
			performer.ClearScreen()
		case 3:
			performer.ClearScrollbackLines()
		default:
			return nil
		}
		return output

	case 'K':
		mode := parameter(params, 0, 0)
		switch mode {
		case 0:
			performer.ClearToLineEnd()
		case 1:
			performer.ClearToLineStart()
		case 2:
			performer.ClearLine()
		default:
			return nil
		}
		return output

	case 'L':
		count := parameter(params, 0, 1)
		performer.InsertLines(count)
		return output

	case 'M':
		count := parameter(params, 0, 1)
		performer.DeleteLines(count)
		return output

	case 'S':
		distance := parameter(params, 0, 1)
		performer.ScrollUp(distance)
		return output

	case 'T':
		distance := parameter(params, 0, 1)
		performer.ScrollDown(distance)
		return output

	case 'X':
		distance := parameter(params, 0, 1)
		performer.ClearCharsAfterCursor(distance)
		return output

	case 'P':
		distance := parameter(params, 0, 1)
		performer.DeleteCharsAfterCursor(distance)
		return output

	case ' ':
		if len(output) > 0 && output[0] == 'q' {
			// Set cursor shape, ignored
			return output[1:]
		}
		return nil

	case 't':
		// Xterm window controls, ignored
		return output

	case 'r':
		top := parameter(params, 0, 1) - 1
		bottom := parameter(params, 1, int(^uint(0)>>1)) - 1
		if top < 0 {
			top = 0
		}
		if bottom < 0 {
			bottom = 0
		}
		performer.SetScrollRegion(top, bottom)
		return output

	case 'n':
		if len(params) > 0 && params[0] == 6 {
			performer.QueryDeviceStatus()
			return output
		}
		return nil

	case 'c':
		performer.QueryTerminalId()
		return output
	}

	return nil
}

func parseFormatting(params []int, performer Performer) {
	if len(params) == 0 {
		params = []int{0}
	}

	for len(params) > 0 {
		param := params[0]
		params = params[1:]

		switch param {
		case 0:
			performer.ResetFormatting()
		case 1:
			performer.SetColorsBright(true)
		case 7:
			performer.SetColorsSwapped(true)
		case 22:
			performer.SetColorsBright(false)
		case 27:
			performer.SetColorsSwapped(false)
		case 30:
			performer.SetForegroundColor(Background)
		case 31:
			performer.SetForegroundColor(Red)
		case 32:
			performer.SetForegroundColor(Green)
		case 33:
			performer.SetForegroundColor(Yellow)
		case 34:
			performer.SetForegroundColor(Blue)
		case 35:
			performer.SetForegroundColor(Magenta)
		case 36:
			performer.SetForegroundColor(Cyan)
		case 37:
			performer.SetForegroundColor(Foreground)
		case 38:
			if color, remaining := parseColorFromParameters(params); remaining != nil {
				performer.SetForegroundColor(color)
				params = remaining
			}
		case 39:
			performer.SetForegroundColor(Foreground)
		case 40:
			performer.SetBackgroundColor(Background)
		case 41:
			performer.SetBackgroundColor(Red)
		case 42:
			performer.SetBackgroundColor(Green)
		case 43:
			performer.SetBackgroundColor(Yellow)
		case 44:
			performer.SetBackgroundColor(Blue)
		case 45:
			performer.SetBackgroundColor(Magenta)
		case 46:
			performer.SetBackgroundColor(Cyan)
		case 47:
			performer.SetBackgroundColor(Foreground)
		case 48:
			if color, remaining := parseColorFromParameters(params); remaining != nil {
				performer.SetBackgroundColor(color)
				params = remaining
			}
		case 49:
			performer.SetBackgroundColor(Background)
		case 90:
			performer.SetForegroundColor(BrightBackground)
		case 91:
			performer.SetForegroundColor(BrightRed)
		case 92:
			performer.SetForegroundColor(BrightGreen)
		case 93:
			performer.SetForegroundColor(BrightYellow)
		case 94:
			performer.SetForegroundColor(BrightBlue)
		case 95:
			performer.SetForegroundColor(BrightMagenta)
		case 96:
			performer.SetForegroundColor(BrightCyan)
		case 97:
			performer.SetForegroundColor(BrightForeground)
		case 100:
			performer.SetBackgroundColor(BrightBackground)
		case 101:
			performer.SetBackgroundColor(BrightRed)
		case 102:
			performer.SetBackgroundColor(BrightGreen)
		case 103:
			performer.SetBackgroundColor(BrightYellow)
		case 104:
			performer.SetBackgroundColor(BrightBlue)
		case 105:
			performer.SetBackgroundColor(BrightMagenta)
		case 106:
			performer.SetBackgroundColor(BrightCyan)
		case 107:
			performer.SetBackgroundColor(BrightForeground)
		}
	}
}

func parseEscapeSequencesOSC(output []byte, performer Performer) []byte {
	var kind int
	kind, output = parseNumericParameter(output)

	if len(output) == 0 || output[0] != ';' {
		return nil
	}

	output = output[1:]

	switch kind {
	case 0, 2:
		var title []byte
		title, output = consumeTerminatedString(output)
		if output == nil {
			return nil
		}

		if len(title) > 0 {
			performer.SetTitle(title)
		} else {
			performer.ResetTitle()
		}
		return output

	case 10, 11:
		if len(output) == 0 || output[0] != '?' {
			return nil
		}
		output = output[1:]

		output = consumeStringTerminator(output)
		if output == nil {
			return nil
		}

		if kind == 10 {
			performer.QueryForegroundColor()
		} else {
			performer.QueryBackgroundColor()
		}
		return output

	default:
		_, output = consumeTerminatedString(output)
		return output
	}
}

func consumeTerminatedString(output []byte) ([]byte, []byte) {
	stringBytes := output

	for len(output) > 0 {
		if remaining := consumeStringTerminator(output); remaining != nil {
			stringLen := len(stringBytes) - len(output)
			return stringBytes[:stringLen], remaining
		}

		output = output[1:]
	}

	return nil, nil
}

func consumeStringTerminator(output []byte) []byte {
	if len(output) > 0 && output[0] == 0x7 {
		return output[1:]
	}
	if len(output) > 1 && output[0] == 0x1B && output[1] == '\\' {
		return output[2:]
	}
	return nil
}

func parseColorFromParameters(params []int) (uint32, []int) {
	if len(params) == 0 {
		return 0, nil
	}

	kind := params[0]
	params = params[1:]

	switch kind {
	case 2:
		// RGB true color
		if len(params) < 3 {
			return 0, nil
		}

		r := params[0] & 0xFF
		g := params[1] & 0xFF
		b := params[2] & 0xFF

		return uint32((r << 8) | (g << 4) | b), params[3:]

	case 5:
		// 256 color table
		if len(params) < 1 {
			return 0, nil
		}
		index := params[0] & 0xFF

		return colorTable[index], params[1:]

	default:
		return 0, nil
	}
}

func parseNumericParameters(output []byte) ([]int, []byte) {
	var params []int

	for {
		var param int
		param, nextOutput := parseNumericParameter(output)
		if nextOutput == nil {
			break
		}

		output = nextOutput

		params = append(params, param)

		if len(output) > 0 && output[0] == ';' {
			output = output[1:]
		} else {
			break
		}
	}

	return params, output
}

func parseNumericParameter(output []byte) (int, []byte) {
	if len(output) == 0 || output[0] < '0' || output[0] > '9' {
		return 0, nil
	}

	param := 0

	for len(output) > 0 && output[0] >= '0' && output[0] <= '9' {
		param = param*10 + int(output[0]-'0')
		output = output[1:]
	}

	return param, output
}

func parameter(params []int, index, defaultVal int) int {
	if index < len(params) {
		return params[index]
	}
	return defaultVal
}

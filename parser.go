package main

var colorTable = [256]uint32{
	0x000000, 0x800000, 0x008000, 0x808000, 0x000080, 0x800080, 0x008080, 0xC0C0C0, 0x808080,
	0xFF0000, 0x00FF00, 0xFFFF00, 0x0000FF, 0xFF00FF, 0x00FFFF, 0xFFFFFF, 0x000000, 0x00005F,
	0x000087, 0x0000AF, 0x0000D7, 0x0000FF, 0x005F00, 0x005F5F, 0x005F87, 0x005FAF, 0x005FD7,
	0x005FFF, 0x008700, 0x00875F, 0x008787, 0x0087AF, 0x0087D7, 0x0087FF, 0x00AF00, 0x00AF5F,
	0x00AF87, 0x00AFAF, 0x00AFD7, 0x00AFFF, 0x00D700, 0x00D75F, 0x00D787, 0x00D7AF, 0x00D7D7,
	0x00D7FF, 0x00FF00, 0x00FF5F, 0x00FF87, 0x00FFAF, 0x00FFD7, 0x00FFFF, 0x5F0000, 0x5F005F,
	0x5F0087, 0x5F00AF, 0x5F00D7, 0x5F00FF, 0x5F5F00, 0x5F5F5F, 0x5F5F87, 0x5F5FAF, 0x5F5FD7,
	0x5F5FFF, 0x5F8700, 0x5F875F, 0x5F8787, 0x5F87AF, 0x5F87D7, 0x5F87FF, 0x5FAF00, 0x5FAF5F,
	0x5FAF87, 0x5FAFAF, 0x5FAFD7, 0x5FAFFF, 0x5FD700, 0x5FD75F, 0x5FD787, 0x5FD7AF, 0x5FD7D7,
	0x5FD7FF, 0x5FFF00, 0x5FFF5F, 0x5FFF87, 0x5FFFAF, 0x5FFFD7, 0x5FFFFF, 0x870000, 0x87005F,
	0x870087, 0x8700AF, 0x8700D7, 0x8700FF, 0x875F00, 0x875F5F, 0x875F87, 0x875FAF, 0x875FD7,
	0x875FFF, 0x878700, 0x87875F, 0x878787, 0x8787AF, 0x8787D7, 0x8787FF, 0x87AF00, 0x87AF5F,
	0x87AF87, 0x87AFAF, 0x87AFD7, 0x87AFFF, 0x87D700, 0x87D75F, 0x87D787, 0x87D7AF, 0x87D7D7,
	0x87D7FF, 0x87FF00, 0x87FF5F, 0x87FF87, 0x87FFAF, 0x87FFD7, 0x87FFFF, 0xAF0000, 0xAF005F,
	0xAF0087, 0xAF00AF, 0xAF00D7, 0xAF00FF, 0xAF5F00, 0xAF5F5F, 0xAF5F87, 0xAF5FAF, 0xAF5FD7,
	0xAF5FFF, 0xAF8700, 0xAF875F, 0xAF8787, 0xAF87AF, 0xAF87D7, 0xAF87FF, 0xAFAF00, 0xAFAF5F,
	0xAFAF87, 0xAFAFAF, 0xAFAFD7, 0xAFAFFF, 0xAFD700, 0xAFD75F, 0xAFD787, 0xAFD7AF, 0xAFD7D7,
	0xAFD7FF, 0xAFFF00, 0xAFFF5F, 0xAFFF87, 0xAFFFAF, 0xAFFFD7, 0xAFFFFF, 0xD70000, 0xD7005F,
	0xD70087, 0xD700AF, 0xD700D7, 0xD700FF, 0xD75F00, 0xD75F5F, 0xD75F87, 0xD75FAF, 0xD75FD7,
	0xD75FFF, 0xD78700, 0xD7875F, 0xD78787, 0xD787AF, 0xD787D7, 0xD787FF, 0xD7AF00, 0xD7AF5F,
	0xD7AF87, 0xD7AFAF, 0xD7AFD7, 0xD7AFFF, 0xD7D700, 0xD7D75F, 0xD7D787, 0xD7D7AF, 0xD7D7D7,
	0xD7D7FF, 0xD7FF00, 0xD7FF5F, 0xD7FF87, 0xD7FFAF, 0xD7FFD7, 0xD7FFFF, 0xFF0000, 0xFF005F,
	0xFF0087, 0xFF00AF, 0xFF00D7, 0xFF00FF, 0xFF5F00, 0xFF5F5F, 0xFF5F87, 0xFF5FAF, 0xFF5FD7,
	0xFF5FFF, 0xFF8700, 0xFF875F, 0xFF8787, 0xFF87AF, 0xFF87D7, 0xFF87FF, 0xFFAF00, 0xFFAF5F,
	0xFFAF87, 0xFFAFAF, 0xFFAFD7, 0xFFAFFF, 0xFFD700, 0xFFD75F, 0xFFD787, 0xFFD7AF, 0xFFD7D7,
	0xFFD7FF, 0xFFFF00, 0xFFFF5F, 0xFFFF87, 0xFFFFAF, 0xFFFFD7, 0xFFFFFF, 0x080808, 0x121212,
	0x1C1C1C, 0x262626, 0x303030, 0x3A3A3A, 0x444444, 0x4E4E4E, 0x585858, 0x626262, 0x6C6C6C,
	0x767676, 0x808080, 0x8A8A8A, 0x949494, 0x9E9E9E, 0xA8A8A8, 0xB2B2B2, 0xBCBCBC, 0xC6C6C6,
	0xD0D0D0, 0xDADADA, 0xE4E4E4, 0xEEEEEE,
}

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

const (
	Background       = 0x01000000
	Foreground       = 0x02000000
	Red              = 0x03000000
	Green            = 0x04000000
	Yellow           = 0x05000000
	Blue             = 0x06000000
	Magenta          = 0x07000000
	Cyan             = 0x08000000
	BrightBackground = 0x09000000
	BrightForeground = 0x0A000000
	BrightRed        = 0x0B000000
	BrightGreen      = 0x0C000000
	BrightYellow     = 0x0D000000
	BrightBlue       = 0x0E000000
	BrightMagenta    = 0x0F000000
	BrightCyan       = 0x10000000
)

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
	if output == nil {
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

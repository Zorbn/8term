package main

import (
	"bytes"
	"fmt"
	"log"

	"github.com/danielgatis/go-vte"
)

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

const (
	Background       uint32 = 0x01000000
	Foreground       uint32 = 0x02000000
	Red              uint32 = 0x03000000
	Green            uint32 = 0x04000000
	Yellow           uint32 = 0x05000000
	Blue             uint32 = 0x06000000
	Magenta          uint32 = 0x07000000
	Cyan             uint32 = 0x08000000
	BrightBackground uint32 = 0x09000000
	BrightForeground uint32 = 0x0A000000
	BrightRed        uint32 = 0x0B000000
	BrightGreen      uint32 = 0x0C000000
	BrightYellow     uint32 = 0x0D000000
	BrightBlue       uint32 = 0x0E000000
	BrightMagenta    uint32 = 0x0F000000
	BrightCyan       uint32 = 0x10000000
)

type grid struct {
	runes            []rune
	foregroundColors []uint32
	backgroundColors []uint32
}

func newGrid() grid {
	runes := make([]rune, emulatorRows*emulatorCols)

	for i := range len(runes) {
		runes[i] = ' '
	}

	foregroundColors := make([]uint32, emulatorRows*emulatorCols)
	backgroundColors := make([]uint32, emulatorRows*emulatorCols)

	return grid{
		runes,
		foregroundColors,
		backgroundColors,
	}
}

const emulatorRows int = 24
const emulatorCols int = 80

type emulator struct {
	grid                              grid
	otherGrid                         grid
	usedHeight                        int
	isInAlternateBuffer               bool
	cursorX, cursorY                  int
	foregroundColor, backgroundColor  uint32
	areColorsBright, areColorsSwapped bool
	input                             bytes.Buffer
}

func newEmulator() emulator {
	grid := newGrid()
	otherGrid := newGrid()

	usedHeight := 1
	isInAlternateBuffer := false
	cursorX, cursorY := 0, 0
	foregroundColor, backgroundColor := Foreground, Background
	areColorsBright, areColorsSwapped := false, false

	var input bytes.Buffer

	return emulator{
		grid,
		otherGrid,
		usedHeight,
		isInAlternateBuffer,
		cursorX, cursorY,
		foregroundColor, backgroundColor,
		areColorsBright, areColorsSwapped,
		input,
	}
}

func (e *emulator) writeRune(r rune) {
	if e.cursorX >= emulatorCols {
		e.cursorX = 0
		e.newlineCursor()
	}

	e.setRune(r, e.cursorX, e.cursorY)
	e.cursorX++
	e.usedHeight = max(e.usedHeight, e.cursorY+1)
}

func (e *emulator) setRune(r rune, x int, y int) {
	index := y*emulatorCols + x

	e.grid.runes[index] = r

	foregroundColor, backgroundColor := e.foregroundColor, e.backgroundColor

	if e.areColorsSwapped {
		foregroundColor, backgroundColor = backgroundColor, foregroundColor
	}

	if e.areColorsBright {
		foregroundColor = brightenTerminalColor(foregroundColor)
		backgroundColor = brightenTerminalColor(backgroundColor)
	}

	e.grid.foregroundColors[index] = foregroundColor
	e.grid.backgroundColors[index] = backgroundColor
}

func (e *emulator) newlineCursor() {
	e.cursorY++

	if e.cursorY >= emulatorRows && !e.isInAlternateBuffer {
		e.scrollContentUp()
	}

	e.cursorY = min(e.cursorY, emulatorRows-1)
}

func (e *emulator) scrollContentUp() {
	copy(e.grid.runes[0:], e.grid.runes[emulatorCols:])
}

func (e *emulator) Print(r rune) {
	e.writeRune(r)
}

func (e *emulator) Execute(b byte) {
	switch b {
	case '\n':
		e.newlineCursor()
	case '\r':
		e.cursorX = 0
	case '\b':
		e.cursorX = max(e.cursorX-1, 0)
	case '\t':
		nextTabStop := (e.cursorX/8 + 1) * 8

		for e.cursorX < nextTabStop {
			e.writeRune(' ')
		}
	case '\a':
		// Nope, not ringing the bell.
	default:
		if isDebug {
			log.Println("Unhandled byte to execute", b)
		}
	}
}

func (e *emulator) Put(b byte) {
	if isDebug {
		fmt.Printf("[Put] %02x\n", b)

	}
}

func (e *emulator) Unhook() {
	if isDebug {
		fmt.Printf("[Unhook]\n")

	}
}

func (e *emulator) Hook(params [][]uint16, intermediates []byte, ignore bool, r rune) {
	if isDebug {
		fmt.Printf("[Hook] params=%v, intermediates=%v, ignore=%v, r=%c\n", params, intermediates, ignore, r)

	}
}

func (e *emulator) OscDispatch(params [][]byte, bellTerminated bool) {
	if isDebug {
		fmt.Printf("[OscDispatch] params=%v, bellTerminated=%v\n", params, bellTerminated)

	}
}

func (e *emulator) CsiDispatch(params [][]uint16, intermediates []byte, ignore bool, r rune) {
	switch r {
	case 'm':
		e.parseFormatting(params)
	case 'l':
		for i := range params {
			param := getParam(params, i, 0)

			switch param {
			case 1047, 1049:
				if e.isInAlternateBuffer {
					e.isInAlternateBuffer = false
					e.grid, e.otherGrid = e.otherGrid, e.grid
				}
			}
		}
	case 'h':
		for i := range params {
			param := getParam(params, i, 0)

			switch param {
			case 1047, 1049:
				if !e.isInAlternateBuffer {
					e.isInAlternateBuffer = true
					e.grid, e.otherGrid = e.otherGrid, e.grid
				}
			}
		}
	case 'A':
		e.cursorY = max(e.cursorY-getParam(params, 0, 1), 0)
	case 'B':
		e.cursorY = min(e.cursorY+getParam(params, 0, 1), emulatorCols-1)
	case 'C':
		e.cursorX = min(e.cursorX+getParam(params, 0, 1), emulatorRows-1)
	case 'D':
		e.cursorX = max(e.cursorX-getParam(params, 0, 1), 0)
	case 'E':
		e.cursorY = min(e.cursorY+getParam(params, 0, 1), emulatorCols-1)
		e.cursorX = 0
	case 'F':
		e.cursorY = max(e.cursorY-getParam(params, 0, 1), 0)
		e.cursorX = 0
	case 'H':
		e.cursorY = getRowsParam(params, 0, 1)
		e.cursorX = getColsParam(params, 1, 1)
	case 'K':
		startX := 0
		endX := emulatorCols

		switch getParam(params, 0, 0) {
		case 0:
			startX = e.cursorX
		case 1:
			endX = e.cursorX
		}

		for x := startX; x < endX; x++ {
			e.setRune(' ', x, e.cursorY)
		}
	case 'J':
		startIndex := 0
		endIndex := emulatorRows * emulatorCols
		cursorIndex := e.cursorY*emulatorCols + e.cursorX

		switch getParam(params, 0, 0) {
		case 0:
			startIndex = cursorIndex
		case 1:
			endIndex = cursorIndex
		case 3:
			// Clear scrollback lines.
			return
		}

		for i := startIndex; i < endIndex; i++ {
			e.grid.runes[i] = ' '
		}
	case 'X':
		startX := e.cursorX
		endX := min(e.cursorX+getParam(params, 0, 1), emulatorCols)

		for x := startX; x < endX; x++ {
			e.setRune(' ', x, e.cursorY)
		}
	default:
		if isDebug {
			log.Printf("Unhandled CSI params=%v, intermediates=%v, ignore=%v, r=%c\n", params, intermediates, ignore, r)
		}
	}
}

func (e *emulator) EscDispatch(intermediates []byte, ignore bool, b byte) {
	if isDebug {
		fmt.Printf("[EscDispatch] intermediates=%v, ignore=%v, byte=%02x\n", intermediates, ignore, b)

	}
}

func (e *emulator) SosPmApcDispatch(kind vte.SosPmApcKind, data []byte, bellTerminated bool) {
	if isDebug {
		kindName := []string{"SOS", "PM", "APC"}[kind]
		fmt.Printf("[SosPmApcDispatch] kind=%s, data=%q, bellTerminated=%v\n", kindName, data, bellTerminated)

	}
}

func getRowsParam(params [][]uint16, index int, def int) int {
	rawValue := getParam(params, index, def)

	return min(max(rawValue, 1), emulatorRows) - 1
}

func getColsParam(params [][]uint16, index int, def int) int {
	rawValue := getParam(params, index, def)

	return min(max(rawValue, 1), emulatorCols) - 1
}

func getParam(params [][]uint16, index int, def int) int {
	if index >= len(params) || len(params[index]) < 1 {
		return def
	}

	return int(params[index][0])
}

func (e *emulator) parseFormatting(params [][]uint16) {
	if len(params) == 0 {
		params = [][]uint16{{0}}
	}

	for i := 0; i < len(params); i++ {
		param := getParam(params, i, 0)

		switch param {
		case 0:
			e.areColorsBright = false
			e.areColorsSwapped = false
			e.foregroundColor = Foreground
			e.backgroundColor = Background
		case 1:
			e.areColorsBright = true
		case 7:
			e.areColorsSwapped = true
		case 22:
			e.areColorsBright = false
		case 27:
			e.areColorsSwapped = false
		case 30:
			e.foregroundColor = Background
		case 31:
			e.foregroundColor = Red
		case 32:
			e.foregroundColor = Green
		case 33:
			e.foregroundColor = Yellow
		case 34:
			e.foregroundColor = Blue
		case 35:
			e.foregroundColor = Magenta
		case 36:
			e.foregroundColor = Cyan
		case 37:
			e.foregroundColor = Foreground
		case 38:
			if color, len := parseColorFromParameters(params, i); len != 0 {
				e.foregroundColor = color
				i += len
			}
		case 39:
			e.foregroundColor = Foreground
		case 40:
			e.backgroundColor = Background
		case 41:
			e.backgroundColor = Red
		case 42:
			e.backgroundColor = Green
		case 43:
			e.backgroundColor = Yellow
		case 44:
			e.backgroundColor = Blue
		case 45:
			e.backgroundColor = Magenta
		case 46:
			e.backgroundColor = Cyan
		case 47:
			e.backgroundColor = Foreground
		case 48:
			if color, len := parseColorFromParameters(params, i); len != 0 {
				e.backgroundColor = color
				i += len
			}
		case 49:
			e.backgroundColor = Background
		case 90:
			e.foregroundColor = BrightBackground
		case 91:
			e.foregroundColor = BrightRed
		case 92:
			e.foregroundColor = BrightGreen
		case 93:
			e.foregroundColor = BrightYellow
		case 94:
			e.foregroundColor = BrightBlue
		case 95:
			e.foregroundColor = BrightMagenta
		case 96:
			e.foregroundColor = BrightCyan
		case 97:
			e.foregroundColor = BrightForeground
		case 100:
			e.backgroundColor = BrightBackground
		case 101:
			e.backgroundColor = BrightRed
		case 102:
			e.backgroundColor = BrightGreen
		case 103:
			e.backgroundColor = BrightYellow
		case 104:
			e.backgroundColor = BrightBlue
		case 105:
			e.backgroundColor = BrightMagenta
		case 106:
			e.backgroundColor = BrightCyan
		case 107:
			e.backgroundColor = BrightForeground
		}
	}
}

func parseColorFromParameters(params [][]uint16, start int) (uint32, int) {
	if start+1 >= len(params) {
		return 0, 0
	}

	kind := getParam(params, start+1, 0)

	switch kind {
	case 2:
		// RGB true color
		if start+4 >= len(params) {
			return 0, 0
		}

		r := getParam(params, start+2, 0) & 0xFF
		g := getParam(params, start+3, 0) & 0xFF
		b := getParam(params, start+4, 0) & 0xFF

		return uint32((r << 16) | (g << 8) | b), 4

	case 5:
		// 256 color table
		if start+2 >= len(params) {
			return 0, 0
		}

		index := getParam(params, start+2, 0) & 0xFF

		return colorTable[index], 2

	default:
		return 0, 0
	}
}

func brightenTerminalColor(color uint32) uint32 {
	switch color {
	case Background:
		return BrightBackground
	case Foreground:
		return BrightForeground
	case Red:
		return BrightRed
	case Green:
		return BrightGreen
	case Yellow:
		return BrightYellow
	case Blue:
		return BrightBlue
	case Magenta:
		return BrightMagenta
	case Cyan:
		return BrightCyan
	default:
		return color
	}
}

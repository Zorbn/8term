package main

import (
	"fmt"
	"log"

	"github.com/danielgatis/go-vte"
)

type grid struct {
	runes []rune
}

func newGrid() grid {
	runes := make([]rune, emulatorRows*emulatorCols)

	for i := range len(runes) {
		runes[i] = ' '
	}

	return grid{
		runes,
	}
}

const emulatorRows int = 24
const emulatorCols int = 80

type emulator struct {
	grid                grid
	otherGrid           grid
	usedHeight          int
	isInAlternateBuffer bool
	cursorX, cursorY    int
}

func newEmulator() emulator {
	grid := newGrid()
	otherGrid := newGrid()

	usedHeight := 1
	isInAlternateBuffer := false
	cursorX, cursorY := 0, 0

	return emulator{
		grid,
		otherGrid,
		usedHeight,
		isInAlternateBuffer,
		cursorX,
		cursorY,
	}
}

func (e *emulator) writeRune(r rune) {
	if e.cursorX >= emulatorCols {
		e.cursorX = 0
		e.cursorY++
	}

	if e.cursorY >= emulatorRows {
		e.scrollContentUp()
		e.cursorY--
	}

	if r == '~' {
		fmt.Println("x, y", e.cursorX, e.cursorY)
	}

	e.setRune(r, e.cursorX, e.cursorY)
	e.cursorX++
	e.usedHeight = max(e.usedHeight, e.cursorY+1)
}

func (e *emulator) setRune(r rune, x int, y int) {
	e.grid.runes[y*emulatorCols+x] = r
}

func (e *emulator) newlineCursor() {
	e.cursorY++

	if e.cursorY >= emulatorRows {
		e.scrollContentUp()
		e.cursorY--
	}
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
		// We'll handle formatting later.
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

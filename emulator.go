package main

import (
	"fmt"
	"log"

	"github.com/danielgatis/go-vte"
)

const emulatorRows int = 24
const emulatorCols int = 80

type emulator struct {
	grid             []rune
	usedHeight       int
	cursorX, cursorY int
}

func newEmulator() emulator {
	grid := make([]rune, emulatorRows*emulatorCols)

	for i := range len(grid) {
		grid[i] = ' '
	}

	usedHeight := 1
	cursorX, cursorY := 0, 0

	return emulator{
		grid,
		usedHeight,
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

	e.grid[e.cursorY*emulatorCols+e.cursorX] = r
	e.cursorX++
	e.usedHeight = max(e.usedHeight, e.cursorY+1)
}

func (e *emulator) newlineCursor() {
	e.cursorY++

	if e.cursorY >= emulatorRows {
		e.scrollContentUp()
		e.cursorY--
	}
}

func (e *emulator) scrollContentUp() {
	copy(e.grid[0:], e.grid[emulatorCols:])
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

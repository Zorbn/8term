package main

import (
	"bytes"
	"unicode/utf8"
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
	input               bytes.Buffer
}

func newEmulator() emulator {
	grid := newGrid()
	otherGrid := newGrid()

	usedHeight := 1
	isInAlternateBuffer := false
	cursorX, cursorY := 0, 0

	var input bytes.Buffer

	return emulator{
		grid,
		otherGrid,
		usedHeight,
		isInAlternateBuffer,
		cursorX,
		cursorY,
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
	e.grid.runes[y*emulatorCols+x] = r
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

func (e *emulator) clampCursorX() {
	e.cursorX = min(max(e.cursorX, 0), emulatorCols-1)
}

func (e *emulator) clampCursorY() {
	e.cursorY = min(max(e.cursorY, 0), emulatorRows-1)
}

func (e *emulator) Plain(text []byte) {
	for i := 0; i < len(text); {
		r, size := utf8.DecodeRune(text[i:])
		i += size

		e.writeRune(r)
	}
}

func (e *emulator) Backspace() {
	e.cursorX--
	e.clampCursorX()
}

func (e *emulator) Tab() {
	nextTabStop := (e.cursorX/8 + 1) * 8

	for e.cursorX < nextTabStop {
		e.writeRune(' ')
	}
}

func (e *emulator) CarriageReturn() {
	e.cursorX = 0
}

func (e *emulator) Newline() {
	e.newlineCursor()
}

func (e *emulator) ReverseNewline() {

}

func (e *emulator) HideCursor() {

}

func (e *emulator) ShowCursor() {

}

func (e *emulator) SwitchToNormalBuffer() {

}

func (e *emulator) SwitchToAlternateBuffer() {

}

func (e *emulator) QueryModifyKeyboard() {

}

func (e *emulator) QueryModifyCursorKeys() {

}

func (e *emulator) QueryModifyFunctionKeys() {

}

func (e *emulator) QueryModifyOtherKeys() {

}

func (e *emulator) QueryDeviceAttributes() {

}

func (e *emulator) ResetFormatting() {

}

func (e *emulator) SetColorsBright(bright bool) {

}

func (e *emulator) SetColorsSwapped(swapped bool) {

}

func (e *emulator) SetForegroundColor(color uint32) {

}

func (e *emulator) SetBackgroundColor(color uint32) {

}

func (e *emulator) SetCursorX(x int) {
	e.cursorX = x
	e.clampCursorX()
}

func (e *emulator) SetCursorY(y int) {
	e.cursorY = y
	e.clampCursorY()
}

func (e *emulator) SetCursorPosition(x, y int) {
	e.cursorX = x
	e.cursorY = y

	e.clampCursorX()
	e.clampCursorY()
}

func (e *emulator) MoveCursorX(delta int) {
	e.cursorX += delta
	e.clampCursorX()
}

func (e *emulator) MoveCursorY(delta int) {
	e.cursorY += delta
	e.clampCursorY()
}

func (e *emulator) MoveCursorYAndResetX(delta int) {
	e.cursorY += delta
	e.cursorX = 0

	e.clampCursorY()
}

func (e *emulator) ClearToScreenEnd() {
	startIndex := e.cursorY*emulatorCols + e.cursorX
	endIndex := emulatorRows * emulatorCols

	for i := startIndex; i < endIndex; i++ {
		e.grid.runes[i] = ' '
	}
}

func (e *emulator) ClearToScreenStart() {
	startIndex := 0
	endIndex := e.cursorY*emulatorCols + e.cursorX

	for i := startIndex; i < endIndex; i++ {
		e.grid.runes[i] = ' '
	}
}

func (e *emulator) ClearScreen() {
	startIndex := 0
	endIndex := emulatorRows * emulatorCols

	for i := startIndex; i < endIndex; i++ {
		e.grid.runes[i] = ' '
	}
}

func (e *emulator) ClearScrollbackLines() {

}

func (e *emulator) ClearToLineEnd() {
	startX := e.cursorX
	endX := emulatorCols

	for x := startX; x < endX; x++ {
		e.setRune(' ', x, e.cursorY)
	}
}

func (e *emulator) ClearToLineStart() {
	startX := 0
	endX := e.cursorX

	for x := startX; x < endX; x++ {
		e.setRune(' ', x, e.cursorY)
	}
}

func (e *emulator) ClearLine() {
	startX := 0
	endX := emulatorCols

	for x := startX; x < endX; x++ {
		e.setRune(' ', x, e.cursorY)
	}
}

func (e *emulator) InsertLines(count int) {

}

func (e *emulator) DeleteLines(count int) {

}

func (e *emulator) ScrollUp(count int) {

}

func (e *emulator) ScrollDown(count int) {

}

func (e *emulator) ClearCharsAfterCursor(count int) {

}

func (e *emulator) DeleteCharsAfterCursor(count int) {

}

func (e *emulator) SetScrollRegion(top, bottom int) {

}

func (e *emulator) QueryDeviceStatus() {

}

func (e *emulator) QueryTerminalId() {

}

func (e *emulator) SetTitle(title []byte) {

}

func (e *emulator) ResetTitle() {

}

func (e *emulator) QueryForegroundColor() {

}

func (e *emulator) QueryBackgroundColor() {

}

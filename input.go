package main

import (
	"slices"
	"unicode/utf8"

	"github.com/Zyko0/go-sdl3/sdl"
)

func handleKeyPress(key sdl.Keycode, focusedPaneIndex *int, panes *[]*pane, command *command, errorFlashTimer *float32, rows, cols int) {
	modState := sdl.GetModState()
	isCmdPressed := (modState & sdl.KMOD_GUI) != 0
	isCtrlPressed := (modState & sdl.KMOD_CTRL) != 0
	isShiftPressed := (modState & sdl.KMOD_SHIFT) != 0

	if isCmdPressed {
		switch key {
		case sdl.K_UP:
			handleMove(panes, focusedPaneIndex, -1, isShiftPressed)
		case sdl.K_DOWN:
			handleMove(panes, focusedPaneIndex, 1, isShiftPressed)
		case sdl.K_X:
			if *focusedPaneIndex < len(*panes) {
				*panes = slices.Delete(*panes, *focusedPaneIndex, *focusedPaneIndex+1)
			}
		}
		return
	}

	if *focusedPaneIndex >= len(*panes) {
		switch key {
		case sdl.K_BACKSPACE:
			command.pop()
		case sdl.K_TAB:
			command.applyCompletion()
		case sdl.K_UP:
			command.historyUp()
		case sdl.K_DOWN:
			command.historyDown()
		case sdl.K_LEFT:
			command.moveCursorLeft()
		case sdl.K_RIGHT:
			command.moveCursorRight()
		case sdl.K_RETURN:
			if runCommand(command, panes, focusedPaneIndex, rows, cols) {
				command.addToHistory()
				command.clear()
			} else {
				*errorFlashTimer = 1
			}
		}

		return
	}

	pane := (*panes)[*focusedPaneIndex]

	if isCtrlPressed && (key >= sdl.K_A && key <= sdl.K_Z) {
		pane.pty.write([]byte{byte(key) & 0x1f})
		return
	}

	switch key {
	case sdl.K_BACKSPACE:
		writeRuneToPty(&pane.pty, '\x7f')
	case sdl.K_TAB:
		writeRuneToPty(&pane.pty, '\t')
	case sdl.K_RETURN:
		writeRuneToPty(&pane.pty, '\r')
	case sdl.K_ESCAPE:
		writeRuneToPty(&pane.pty, '\x1b')
	}
}

func handleMove(panes *[]*pane, focusedPaneIndex *int, dir int, isShiftPressed bool) {
	startPaneIndex := *focusedPaneIndex

	for needsMove := true; shouldMove(needsMove, *panes, *focusedPaneIndex, dir); needsMove = false {
		*focusedPaneIndex += dir
	}

	if isShiftPressed {
		swapPanes(*panes, *focusedPaneIndex, startPaneIndex)
	}
}

func shouldMove(needsMove bool, panes []*pane, focusedPaneIndex int, dir int) bool {
	canMove := (dir > 0 && focusedPaneIndex < len(panes)) || (dir < 0 && focusedPaneIndex > 0)

	if !canMove {
		return false
	}

	if needsMove {
		return true
	}

	isPaneEmpty := panes[focusedPaneIndex].emulator.grid.usedHeight == 0

	return isPaneEmpty
}

func swapPanes(panes []*pane, focusedPaneIndex int, startPaneIndex int) {
	if startPaneIndex < len(panes) && focusedPaneIndex < len(panes) {
		panes[focusedPaneIndex], panes[startPaneIndex] = panes[startPaneIndex], panes[focusedPaneIndex]
	}
}

func runCommand(command *command, panes *[]*pane, focusedPaneIndex *int, rows, cols int) bool {
	ast, didSucceed := command.parse()

	if !didSucceed {
		return false
	}

	pane := newPane(rows, cols)
	pane.run(ast)

	*panes = append(*panes, &pane)
	*focusedPaneIndex++

	return true
}

func writeRuneToPty(pty *pty, r rune) {
	var buffer [4]byte

	size := utf8.EncodeRune(buffer[:], r)
	pty.write(buffer[:size])
}

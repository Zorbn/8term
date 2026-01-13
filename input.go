package main

import (
	"os"
	"slices"
	"unicode/utf8"

	"github.com/Zyko0/go-sdl3/sdl"
)

func handleKeyPress(key sdl.Keycode, focusedPaneIndex *int, panes *[]*pane, command *command, errorFlashTimer *float32, homeDir string) {

	modState := sdl.GetModState()
	cmdPressed := (modState & sdl.KMOD_GUI) != 0
	ctrlPressed := (modState & sdl.KMOD_CTRL) != 0

	if cmdPressed {
		switch key {
		case sdl.K_UP:
			for needsMove := true; shouldMove(needsMove, panes, focusedPaneIndex, -1); needsMove = false {
				*focusedPaneIndex--
			}
		case sdl.K_DOWN:
			for needsMove := true; shouldMove(needsMove, panes, focusedPaneIndex, 1); needsMove = false {
				*focusedPaneIndex++
			}
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
		case sdl.K_RETURN:
			tokenizedCommand := command.tokenize()

			if runCommand(tokenizedCommand, panes, focusedPaneIndex, homeDir) {
				command.clear()
			} else {
				*errorFlashTimer = 1
			}
		}

		return
	}

	pane := (*panes)[*focusedPaneIndex]

	if ctrlPressed && (key >= sdl.K_A && key <= sdl.K_Z) {
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

func shouldMove(needsMove bool, panes *[]*pane, focusedPaneIndex *int, dir int) bool {
	canMove := (dir > 0 && *focusedPaneIndex < len(*panes)) || (dir < 0 && *focusedPaneIndex > 0)

	if !canMove {
		return false
	}

	if needsMove {
		return true
	}

	isPaneEmpty := (*panes)[*focusedPaneIndex].emulator.usedHeight == 0

	return isPaneEmpty
}

func runCommand(tokenizedCommand *tokenizeResult, panes *[]*pane, focusedPaneIndex *int, homeDir string) bool {
	var stringTokens []string

	for _, t := range tokenizedCommand.tokens {
		stringTokens = append(stringTokens, string(t))
	}

	if len(tokenizedCommand.tokens) == 0 {
		return false
	}

	switch stringTokens[0] {
	case "cd":
		if len(stringTokens) > 2 {
			return false
		}

		path := homeDir

		if len(stringTokens) > 1 {
			path = stringTokens[1]
		}

		os.Chdir(path)
	default:
		pane, err := newPane(stringTokens[0], stringTokens[1:]...)

		if err != nil {
			return false
		}

		pane.run()

		*panes = append(*panes, &pane)
		*focusedPaneIndex++
	}

	return true
}

func writeRuneToPty(pty *pty, r rune) {
	var buffer [4]byte

	size := utf8.EncodeRune(buffer[:], r)
	pty.write(buffer[:size])
}

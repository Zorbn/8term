package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"os"
	"slices"
	"unicode/utf8"
	"unsafe"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var isDebug bool

//go:embed Inconsolata-Regular.ttf
var fontData []byte

//go:embed sdf.fs
var fragmentShaderCode string

func main() {
	isDebugFlag := flag.Bool("debug", false, "Enable debugging mode")
	flag.Parse()
	isDebug = *isDebugFlag

	fmt.Println("Hello, world!", isDebug)

	rl.SetConfigFlags(rl.FlagWindowHighdpi | rl.FlagMsaa4xHint | rl.FlagWindowResizable)

	const screenWidth, screenHeight = 800, 450
	rl.InitWindow(screenWidth, screenHeight, "8term")
	defer rl.CloseWindow()

	rl.SetExitKey(rl.KeyNull)

	dpi := rl.GetWindowScaleDPI().X
	fontSize := int32(16 * dpi)

	font := rl.Font{BaseSize: fontSize, CharsCount: 95}
	defer rl.UnloadFont(font)

	glyphs := rl.LoadFontData(fontData, fontSize, nil, 0, rl.FontSdf)
	font.Chars = &glyphs[0]

	atlas := rl.GenImageFontAtlas(unsafe.Slice(font.Chars, font.CharsCount), unsafe.Slice(&font.Recs, font.CharsCount), fontSize, 0, 1)
	font.Texture = rl.LoadTextureFromImage(&atlas)
	rl.UnloadImage(&atlas)

	shader := rl.LoadShaderFromMemory("", fragmentShaderCode)
	defer rl.UnloadShader(shader)
	rl.SetTextureFilter(font.Texture, rl.FilterBilinear)

	scaledFontSize := float32(fontSize) / dpi
	glyphSize := rl.MeasureTextEx(font, "M", scaledFontSize, 0)

	var cameraY float32 = 0
	// var cameraSpeed float32 = 10
	cameraMargin := glyphSize.Y * 3

	var panes []*pane
	var command []rune
	focusedPaneIndex := 0
	var ptyInputBuffer [4]byte

	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")

	rl.SetTargetFPS(144)

	for !rl.WindowShouldClose() {
		// Update:
		if rl.IsKeyDown(rl.KeyLeftSuper) || rl.IsKeyDown(rl.KeyRightSuper) {
			if isKeyPressedOrRepeated(rl.KeyUp) {
				focusedPaneIndex = max(focusedPaneIndex-1, 0)
			}

			if isKeyPressedOrRepeated(rl.KeyDown) {
				focusedPaneIndex = min(focusedPaneIndex+1, len(panes))
			}

			if isKeyPressedOrRepeated(rl.KeyW) && focusedPaneIndex < len(panes) {
				panes = slices.Delete(panes, focusedPaneIndex, focusedPaneIndex+1)
			}
		}

		if focusedPaneIndex >= len(panes) {
			// Send input to command:
			for {
				r := rune(rl.GetCharPressed())

				if r == 0 {
					break
				}

				command = append(command, r)
			}

			// TODO:
			// Should actually use GetKeyPressed in a loop to get pressed keys then check each frame if pressed keys have been released.
			// Then we would need our own repeat timer logic.

			if isKeyPressedOrRepeated(rl.KeyBackspace) {
				if len(command) > 0 {
					command = command[:len(command)-1]
				}
			}

			if isKeyPressedOrRepeated(rl.KeyEnter) {
				if len(command) > 0 {
					// TODO: This program should be its own shell.
					pane := newPane("bash", "-c", string(command))
					pane.run()

					command = nil

					panes = append(panes, &pane)
					focusedPaneIndex++
				}
			}
		} else {
			// Send input to pane:
			pane := panes[focusedPaneIndex]

			for {
				r := rune(rl.GetCharPressed())

				if r == 0 {
					break
				}

				writeRuneToPty(&pane.pty, ptyInputBuffer, r)
			}

			if isKeyPressedOrRepeated(rl.KeyBackspace) {
				writeRuneToPty(&pane.pty, ptyInputBuffer, '\x7f')
			}

			if isKeyPressedOrRepeated(rl.KeyTab) {
				writeRuneToPty(&pane.pty, ptyInputBuffer, '\t')
			}

			if isKeyPressedOrRepeated(rl.KeyEnter) {
				writeRuneToPty(&pane.pty, ptyInputBuffer, '\r')
			}

			if isKeyPressedOrRepeated(rl.KeyEscape) {
				writeRuneToPty(&pane.pty, ptyInputBuffer, '\x1b')
			}
		}

		var paneY float32 = 0

		for i, pane := range panes {
			pane.handleOutput()

			if i < focusedPaneIndex {
				paneY += glyphSize.Y * float32(pane.emulator.usedHeight+1)
			}
		}

		// dt := rl.GetFrameTime()
		windowHeight := float32(rl.GetRenderHeight()) / dpi

		// cameraY = rl.Lerp(cameraY, paneY-windowHeight+cameraMargin, dt*cameraSpeed)
		focusedPaneHeight := glyphSize.Y

		if focusedPaneIndex < len(panes) {
			focusedPaneHeight = glyphSize.Y * float32(panes[focusedPaneIndex].emulator.usedHeight)
		}

		cameraY = paneY - windowHeight + focusedPaneHeight + cameraMargin

		// Draw:
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.Translatef(0, -cameraY, 0)

		paneWidth := glyphSize.X * float32(emulatorCols)
		paneY = 0

		for i, pane := range panes {
			emulator := &pane.emulator

			paneHeight := glyphSize.Y * float32(emulator.usedHeight)

			if paneY+paneHeight > cameraY {
				color := getPaneColor(i, focusedPaneIndex)
				rl.DrawRectangleV(rl.NewVector2(0, paneY), rl.NewVector2(paneWidth, paneHeight), color)
			}

			paneY += glyphSize.Y * float32(emulator.usedHeight+1)
		}

		rl.BeginShaderMode(shader)

		paneY = 0

		for _, pane := range panes {
			emulator := &pane.emulator

			paneHeight := glyphSize.Y * float32(emulator.usedHeight)

			if paneY+paneHeight > cameraY {
				for y := range emulator.usedHeight {
					lineStartIndex := y * emulatorCols
					lineEndIndex := lineStartIndex + emulatorCols
					line := emulator.grid.runes[lineStartIndex:lineEndIndex]
					lineY := glyphSize.Y*float32(y) + paneY

					rl.DrawTextCodepoints(font, line, rl.NewVector2(0, lineY), scaledFontSize, 0, rl.Black)
				}
			}

			paneY += glyphSize.Y * float32(emulator.usedHeight+1)
		}

		color := getPaneColor(len(panes), focusedPaneIndex)
		rl.DrawRectangleV(rl.NewVector2(0, paneY), rl.NewVector2(paneWidth, glyphSize.Y), color)
		rl.DrawTextCodepoints(font, command, rl.NewVector2(0, paneY), scaledFontSize, 0, rl.Black)

		rl.EndShaderMode()

		rl.EndDrawing()

	}
}

func writeRuneToPty(pty *pty, buffer [4]byte, r rune) {
	size := utf8.EncodeRune(buffer[:], r)
	pty.write(buffer[:size])
}

func getPaneColor(index, focusedIndex int) color.RGBA {
	if index == focusedIndex {
		return rl.SkyBlue
	} else {
		return rl.LightGray
	}
}

func isKeyPressedOrRepeated(key int32) bool {
	return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
}

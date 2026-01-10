package main

import (
	_ "embed"
	"flag"
	"fmt"
	"image/color"
	"os"
	"slices"
	"unicode"
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
	paneBorderWidth := glyphSize.X / 2

	var cameraY float32 = 0
	var cameraSpeed float32 = 10
	cameraMargin := glyphSize.Y * 3

	var panes []*pane
	var command []rune
	focusedPaneIndex := 0
	var ptyInputBuffer [4]byte

	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")

	homeDir, err := os.UserHomeDir()

	if err != nil {
		homeDir = ""
	}

	var errorFlashTimer float32

	// rl.SetTargetFPS(144)

	for !rl.WindowShouldClose() {
		// Update:
		dt := rl.GetFrameTime()
		errorFlashTimer -= dt

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
				if runCommand(command, &panes, &focusedPaneIndex, homeDir) {
					command = nil
				} else {
					errorFlashTimer = 1
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

		windowWidth := float32(rl.GetRenderWidth()) / dpi
		windowHeight := float32(rl.GetRenderHeight()) / dpi

		focusedPaneHeight := glyphSize.Y

		if focusedPaneIndex < len(panes) {
			focusedPaneHeight = glyphSize.Y * float32(panes[focusedPaneIndex].emulator.usedHeight)
		}

		cameraY = rl.Lerp(cameraY, paneY-windowHeight+focusedPaneHeight+cameraMargin, dt*cameraSpeed)
		cameraX := (glyphSize.X*float32(emulatorCols) - windowWidth) / 2

		// Draw:
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		rl.Translatef(-cameraX, -cameraY, 0)

		paneWidth := glyphSize.X * float32(emulatorCols)
		paneY = 0

		for i, pane := range panes {
			emulator := &pane.emulator

			paneHeight := glyphSize.Y * float32(emulator.usedHeight)

			if paneY+paneHeight > cameraY {
				borderColor := getPaneBorderColor(i, focusedPaneIndex)
				drawBorderedRect(rl.NewVector2(0, paneY), rl.NewVector2(paneWidth, paneHeight), paneBorderWidth, borderColor, rl.Black)
			}

			paneY += glyphSize.Y * float32(emulator.usedHeight+1)
		}

		rl.BeginShaderMode(shader)

		paneY = 0

		for paneIndex, pane := range panes {
			emulator := &pane.emulator

			paneHeight := glyphSize.Y * float32(emulator.usedHeight)

			if paneY+paneHeight > cameraY {
				for y := range emulator.usedHeight {
					lineY := glyphSize.Y*float32(y) + paneY

					for x := range emulatorCols {
						i := y*emulatorCols + x
						r := emulator.grid.runes[i]
						foregroundColor := emulator.grid.foregroundColors[i]
						backgroundColor := emulator.grid.backgroundColors[i]

						position := rl.NewVector2(glyphSize.X*float32(x), lineY)

						if backgroundColor != Background {
							color := terminalColorToColor(backgroundColor)
							rl.DrawRectangleV(position, glyphSize, color)
						}

						if !unicode.IsSpace(r) {
							color := terminalColorToColor(foregroundColor)
							rl.DrawTextCodepoint(font, r, position, scaledFontSize, color)
						}
					}
				}

				if paneIndex == focusedPaneIndex && emulator.cursorY < emulator.usedHeight {
					r := emulator.grid.runes[emulator.cursorY*emulatorCols+emulator.cursorX]
					position := rl.NewVector2(glyphSize.X*float32(emulator.cursorX), paneY+glyphSize.Y*float32(emulator.cursorY))

					rl.DrawRectangleV(position, glyphSize, rl.White)
					rl.DrawTextCodepoint(font, r, position, scaledFontSize, rl.Black)
				}
			}

			paneY += glyphSize.Y * float32(emulator.usedHeight+1)
		}

		borderColor := getPaneBorderColor(len(panes), focusedPaneIndex)
		drawBorderedRect(rl.NewVector2(0, paneY), rl.NewVector2(paneWidth, glyphSize.Y), paneBorderWidth, borderColor, rl.Black)

		if errorFlashTimer > 0 {
			errorColor := rl.Red
			errorColor.A = uint8(errorFlashTimer * 255)

			rl.DrawRectangleV(rl.NewVector2(0, paneY), rl.NewVector2(paneWidth, glyphSize.Y), errorColor)
		}

		rl.DrawTextCodepoints(font, command, rl.NewVector2(0, paneY), scaledFontSize, 0, rl.White)

		if len(panes) == focusedPaneIndex {
			position := rl.NewVector2(glyphSize.X*float32(len(command)), paneY)

			rl.DrawRectangleV(position, glyphSize, rl.White)
		}

		rl.EndShaderMode()

		rl.EndDrawing()

	}
}

func runCommand(command []rune, panes *[]*pane, focusedPaneIndex *int, homeDir string) bool {
	tokens := tokenize(command)
	var stringTokens []string

	for _, t := range tokens {
		stringTokens = append(stringTokens, string(t))
	}

	if len(tokens) == 0 {
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

func writeRuneToPty(pty *pty, buffer [4]byte, r rune) {
	size := utf8.EncodeRune(buffer[:], r)
	pty.write(buffer[:size])
}

func drawBorderedRect(position, size rl.Vector2, borderWidth float32, borderColor, backgroundColor color.RGBA) {
	borderOffset := rl.NewVector2(borderWidth, borderWidth)
	borderPosition := rl.Vector2Subtract(position, borderOffset)
	borderSize := rl.Vector2Add(size, rl.Vector2Scale(borderOffset, 2))

	rl.DrawRectangleV(borderPosition, borderSize, borderColor)
	rl.DrawRectangleV(position, size, backgroundColor)
}

func getPaneBorderColor(index, focusedIndex int) color.RGBA {
	if index == focusedIndex {
		return rl.SkyBlue
	} else {
		return rl.LightGray
	}
}

func isKeyPressedOrRepeated(key int32) bool {
	return rl.IsKeyPressed(key) || rl.IsKeyPressedRepeat(key)
}

func terminalColorToColor(color uint32) color.RGBA {
	switch color {
	case Background:
		return rl.Black
	case Foreground:
		return rl.White
	case Red:
		return rl.Red
	case Green:
		return rl.Green
	case Yellow:
		return rl.Yellow
	case Blue:
		return rl.Blue
	case Magenta:
		return rl.Magenta
	case Cyan:
		return rl.SkyBlue
	case BrightBackground:
		return brightenColor(rl.Black)
	case BrightForeground:
		return brightenColor(rl.White)
	case BrightRed:
		return brightenColor(rl.Red)
	case BrightGreen:
		return brightenColor(rl.Green)
	case BrightYellow:
		return brightenColor(rl.Yellow)
	case BrightBlue:
		return brightenColor(rl.Blue)
	case BrightMagenta:
		return brightenColor(rl.Magenta)
	case BrightCyan:
		return brightenColor(rl.SkyBlue)
	default:
		r := (color >> 16) & 0xFF
		g := (color >> 8) & 0xFF
		b := color & 0xFF

		return rl.NewColor(uint8(r), uint8(g), uint8(b), 255)
	}
}

func brightenColor(color color.RGBA) color.RGBA {
	r := brightenColorComponent(color.R)
	g := brightenColorComponent(color.G)
	b := brightenColorComponent(color.B)

	return rl.NewColor(r, g, b, color.A)
}

func brightenColorComponent(x uint8) uint8 {
	return uint8((int(x)*2 + 255) / 3)
}

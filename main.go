package main

import (
	"flag"
	"fmt"
	"image/color"
	"math"
	"os"
	"slices"

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

var isDebug bool

const defaultWindowWidth, defaultWindowHeight = 800, 800

func main() {
	isDebugFlag := flag.Bool("debug", false, "Enable debugging mode")
	flag.Parse()
	isDebug = *isDebugFlag
	fmt.Println("Hello, world!", isDebug)

	defer binsdl.Load().Unload()
	defer binttf.Load().Unload()

	if err := sdl.Init(sdl.INIT_VIDEO); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	if err := ttf.Init(); err != nil {
		panic(err)
	}
	defer ttf.Quit()

	window, renderer, err := sdl.CreateWindowAndRenderer("8term", defaultWindowWidth, defaultWindowHeight, sdl.WINDOW_RESIZABLE|sdl.WINDOW_HIGH_PIXEL_DENSITY)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()
	defer renderer.Destroy()

	window.StartTextInput()
	renderer.SetVSync(1)
	renderer.SetDrawBlendMode(sdl.BLENDMODE_BLEND)

	dpi, err := window.PixelDensity()
	if err != nil {
		panic(err)
	}

	atlas, err := loadGlyphAtlas(renderer, dpi)
	if err != nil {
		panic(err)
	}
	defer atlas.Destroy()

	run(renderer, atlas)
}

func run(renderer *sdl.Renderer, atlas *GlyphAtlas) {
	glyphSize := Vector2{atlas.glyphWidth, atlas.glyphHeight}
	paneBorderWidth := atlas.glyphWidth / 2

	var cameraY float32 = 0
	var cameraSpeed float32 = 10
	cameraMargin := atlas.glyphHeight * 3

	var panes []*pane
	var command command
	focusedPaneIndex := 0

	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	var errorFlashTimer float32

	lastTime := sdl.Ticks()
	var time float32
	running := true

	for running {
		currentTime := sdl.Ticks()
		dt := float32(currentTime-lastTime) / 1000.0
		lastTime = currentTime
		time += dt
		errorFlashTimer -= dt
		didResize := false

		var event sdl.Event
		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				running = false
			case sdl.EVENT_TEXT_INPUT:
				textEvent := event.TextInputEvent()

				if focusedPaneIndex >= len(panes) {
					for _, r := range textEvent.Text {
						command.append(r)
					}
				} else {
					pane := panes[focusedPaneIndex]
					for _, r := range textEvent.Text {
						writeRuneToPty(&pane.pty, r)
					}
				}
			case sdl.EVENT_KEY_DOWN:
				keyEvent := event.KeyboardEvent()
				handleKeyPress(keyEvent.Key, &focusedPaneIndex, &panes, &command, &errorFlashTimer, homeDir)
			case sdl.EVENT_WINDOW_RESIZED:
				didResize = true
			}
		}

		var paneY float32 = 0
		for i, pane := range panes {
			if i < focusedPaneIndex {
				paneY += getPaneHeight(pane, atlas)
			}

			pane.timer += dt
			isRunning := pane.handleOutput()

			if pane.emulator.grid.usedHeight > 0 {
				continue
			}

			if isRunning {
				continue
			}

			panes = slices.Delete(panes, i, i+1)

			if focusedPaneIndex > i {
				focusedPaneIndex--
			}
		}

		sdlWindowWidth, sdlWindowHeight, err := renderer.RenderOutputSize()
		if err != nil {
			panic(err)
		}
		windowWidth := float32(sdlWindowWidth)
		windowHeight := float32(sdlWindowHeight)

		var targetY float32
		if focusedPaneIndex < len(panes) {
			targetY = paneY - (windowHeight-float32(panes[focusedPaneIndex].emulator.grid.usedHeight)*atlas.glyphHeight)/2
		} else {
			targetY = paneY - windowHeight + atlas.glyphHeight + cameraMargin
		}

		if didResize || cameraY == 0 {
			cameraY = targetY
		} else {
			cameraY = lerp(cameraY, targetY, dt*cameraSpeed)
		}

		draw(renderer, atlas, panes, focusedPaneIndex, cameraY, windowWidth, windowHeight, time,
			errorFlashTimer, &command, paneBorderWidth, glyphSize)

		renderer.Present()
	}
}

func draw(renderer *sdl.Renderer, atlas *GlyphAtlas, panes []*pane, focusedPaneIndex int, cameraY, windowWidth, windowHeight, time, errorFlashTimer float32, command *command, paneBorderWidth float32, glyphSize Vector2) {
	cameraX := (atlas.glyphWidth*float32(emulatorCols) - windowWidth) / 2

	backgroundR := uint8(math.Sin(0.3*float64(time)+0)*10 + 10)
	backgroundG := uint8(math.Sin(0.3*float64(time)+2)*10 + 10)
	backgroundB := uint8(math.Sin(0.3*float64(time)+4)*10 + 10)

	renderer.SetDrawColor(backgroundR, backgroundG, backgroundB, 255)
	renderer.Clear()

	paneWidth := atlas.glyphWidth * float32(emulatorCols)

	var paneY float32 = 0

	for i, pane := range panes {
		emulator := &pane.emulator
		paneHeight := atlas.glyphHeight * float32(emulator.grid.usedHeight)

		if isPaneVisible(paneY, paneHeight, cameraY, windowHeight) {
			borderColor := getPaneBorderColor(i, focusedPaneIndex, pane)
			borderWidth := getPaneBorderWidth(i, focusedPaneIndex, paneBorderWidth, pane.timer)
			drawBorderedRect(renderer, atlas, cameraX, cameraY,
				Vector2{0, paneY}, Vector2{paneWidth, paneHeight},
				borderWidth, borderColor, color.RGBA{0, 0, 0, 255})
		}

		paneY += getPaneHeight(pane, atlas)
	}

	paneY = 0

	for paneIndex, pane := range panes {
		emulator := &pane.emulator
		paneHeight := atlas.glyphHeight * float32(emulator.grid.usedHeight)

		if isPaneVisible(paneY, paneHeight, cameraY, windowHeight) {
			for y := range emulator.grid.usedHeight {
				lineY := atlas.glyphHeight*float32(y) + paneY
				for x := range emulatorCols {
					i := y*emulatorCols + x
					r := emulator.grid.runes[i]
					fg := emulator.grid.foregroundColors[i]
					bg := emulator.grid.backgroundColors[i]

					position := Vector2{atlas.glyphWidth * float32(x), lineY}

					if bg != Background {
						c := terminalColorToColor(bg)
						drawRect(renderer, atlas, cameraX, cameraY, position, glyphSize, c)
					}

					c := terminalColorToColor(fg)
					drawGlyph(renderer, atlas, cameraX, cameraY, r, position, c)
				}
			}

			if paneIndex == focusedPaneIndex && emulator.cursorY < emulator.grid.usedHeight {
				r := emulator.grid.runes[emulator.cursorY*emulatorCols+emulator.cursorX]
				position := Vector2{
					atlas.glyphWidth * float32(emulator.cursorX),
					paneY + atlas.glyphHeight*float32(emulator.cursorY),
				}
				drawRect(renderer, atlas, cameraX, cameraY, position, glyphSize,
					color.RGBA{255, 255, 255, 255})
				drawGlyph(renderer, atlas, cameraX, cameraY, r, position,
					color.RGBA{0, 0, 0, 255})
			}
		}
		paneY += getPaneHeight(pane, atlas)
	}

	borderColor := getPaneBorderColor(len(panes), focusedPaneIndex, nil)
	borderWidth := getPaneBorderWidth(len(panes), focusedPaneIndex, paneBorderWidth, time)
	drawBorderedRect(renderer, atlas, cameraX, cameraY,
		Vector2{0, paneY}, Vector2{paneWidth, atlas.glyphHeight},
		borderWidth, borderColor, color.RGBA{0, 0, 0, 255})

	if errorFlashTimer > 0 {
		errorColor := color.RGBA{255, 0, 0, uint8(errorFlashTimer * 255)}
		drawRect(renderer, atlas, cameraX, cameraY,
			Vector2{0, paneY}, Vector2{paneWidth, atlas.glyphHeight}, errorColor)
	}

	cwd, err := os.Getwd()
	if err != nil {
		cwd = "?"
	}

	var commandX float32
	commandX += drawString(renderer, atlas, cameraX, cameraY, cwd,
		Vector2{commandX, paneY}, color.RGBA{255, 255, 255, 255})

	commandX += drawString(renderer, atlas, cameraX, cameraY, "> ",
		Vector2{commandX, paneY}, color.RGBA{255, 255, 255, 255})

	commandX += drawText(renderer, atlas, cameraX, cameraY, command.runes,
		Vector2{commandX, paneY}, color.RGBA{255, 255, 255, 255})

	command.parse()
	missingTrailingRunes := slices.Concat(command.tokenizer.missingTrailingRunes, command.parser.missingTrailingRunes)

	if len(missingTrailingRunes) > 0 {
		drawText(renderer, atlas, cameraX, cameraY,
			missingTrailingRunes,
			Vector2{commandX, paneY},
			color.RGBA{255, 255, 255, 255})
	}

	if len(panes) == focusedPaneIndex {
		position := Vector2{commandX, paneY}
		drawRect(renderer, atlas, cameraX, cameraY, position, glyphSize,
			color.RGBA{255, 255, 255, 255})

		if len(missingTrailingRunes) > 0 {
			drawGlyph(renderer, atlas, cameraX, cameraY,
				missingTrailingRunes[0], position,
				color.RGBA{0, 0, 0, 255})
		}
	}
}

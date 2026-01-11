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

	"github.com/Zyko0/go-sdl3/bin/binsdl"
	"github.com/Zyko0/go-sdl3/bin/binttf"
	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

var isDebug bool

//go:embed Inconsolata-Regular.ttf
var fontData []byte

type Vector2 struct {
	X, Y float32
}

const startGlyph = ' ' + 1

type GlyphAtlas struct {
	texture     *sdl.Texture
	glyphs      []sdl.FRect
	glyphWidth  float32
	glyphHeight float32
}

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

	const screenWidth, screenHeight = 800, 450
	window, renderer, err := sdl.CreateWindowAndRenderer("8term", screenWidth, screenHeight, sdl.WINDOW_RESIZABLE|sdl.WINDOW_HIGH_PIXEL_DENSITY)

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

	fontSize := 16 * dpi

	rwops, err := sdl.IOFromConstMem(fontData)

	if err != nil {
		panic(err)
	}

	font, err := ttf.OpenFontIO(rwops, true, fontSize)

	if err != nil {
		panic(err)
	}

	defer font.Close()

	atlas := createGlyphAtlas(renderer, font)
	defer atlas.texture.Destroy()

	glyphSize := Vector2{atlas.glyphWidth, atlas.glyphHeight}

	paneBorderWidth := atlas.glyphWidth / 2

	var cameraY float32 = 0
	var cameraSpeed float32 = 10
	cameraMargin := atlas.glyphHeight * 3

	var panes []*pane
	var command []rune
	var tokenizedCommand tokenizeResult
	focusedPaneIndex := 0
	var ptyInputBuffer [4]byte

	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")

	homeDir, err := os.UserHomeDir()

	if err != nil {
		homeDir = ""
	}

	var errorFlashTimer float32

	lastTime := sdl.Ticks()
	running := true

	for running {
		currentTime := sdl.Ticks()
		dt := float32(currentTime-lastTime) / 1000.0
		lastTime = currentTime
		errorFlashTimer -= dt

		var event sdl.Event

		for sdl.PollEvent(&event) {
			switch event.Type {
			case sdl.EVENT_QUIT:
				running = false

			case sdl.EVENT_TEXT_INPUT:
				textEvent := event.TextInputEvent()
				if focusedPaneIndex >= len(panes) {
					for _, r := range textEvent.Text {
						if r != 0 {
							command = append(command, r)
						}
					}
				} else {
					pane := panes[focusedPaneIndex]
					for _, r := range textEvent.Text {
						if r != 0 {
							writeRuneToPty(&pane.pty, ptyInputBuffer, r)
						}
					}
				}

			case sdl.EVENT_KEY_DOWN:
				keyEvent := event.KeyboardEvent()

				handleKeyPress(keyEvent.Key, &focusedPaneIndex, &panes, &command, &tokenizedCommand,
					&errorFlashTimer, homeDir, ptyInputBuffer)
			}
		}

		var paneY float32 = 0

		for i, pane := range panes {
			pane.handleOutput()

			if i < focusedPaneIndex {
				paneY += atlas.glyphHeight * float32(pane.emulator.usedHeight+1)
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
			targetY = paneY - (windowHeight-float32(panes[focusedPaneIndex].emulator.usedHeight)*atlas.glyphHeight)/2
		} else {
			targetY = paneY - windowHeight + atlas.glyphHeight + cameraMargin
		}

		cameraY = lerp(cameraY, targetY, dt*cameraSpeed)
		cameraX := (atlas.glyphWidth*float32(emulatorCols) - windowWidth) / 2

		renderer.SetDrawColor(245, 245, 245, 255)
		renderer.Clear()

		paneWidth := atlas.glyphWidth * float32(emulatorCols)
		paneY = 0

		for i, pane := range panes {
			emulator := &pane.emulator

			paneHeight := atlas.glyphHeight * float32(emulator.usedHeight)

			if paneY+paneHeight > cameraY {
				borderColor := getPaneBorderColor(i, focusedPaneIndex)
				drawBorderedRect(renderer, cameraX, cameraY,
					Vector2{0, paneY}, Vector2{paneWidth, paneHeight},
					paneBorderWidth, borderColor, color.RGBA{0, 0, 0, 255})
			}

			paneY += atlas.glyphHeight * float32(emulator.usedHeight+1)
		}

		paneY = 0

		for paneIndex, pane := range panes {
			emulator := &pane.emulator

			paneHeight := atlas.glyphHeight * float32(emulator.usedHeight)

			if paneY+paneHeight > cameraY {
				for y := range emulator.usedHeight {
					lineY := atlas.glyphHeight*float32(y) + paneY

					for x := range emulatorCols {
						i := y*emulatorCols + x
						r := emulator.grid.runes[i]
						foregroundColor := emulator.grid.foregroundColors[i]
						backgroundColor := emulator.grid.backgroundColors[i]

						position := Vector2{atlas.glyphWidth * float32(x), lineY}

						if backgroundColor != Background {
							c := terminalColorToColor(backgroundColor)
							drawRect(renderer, cameraX, cameraY, position, glyphSize, c)
						}

						if !unicode.IsSpace(r) {
							c := terminalColorToColor(foregroundColor)
							drawGlyph(renderer, &atlas, cameraX, cameraY, r, position, c)
						}
					}
				}

				if paneIndex == focusedPaneIndex && emulator.cursorY < emulator.usedHeight {
					r := emulator.grid.runes[emulator.cursorY*emulatorCols+emulator.cursorX]
					position := Vector2{
						atlas.glyphWidth * float32(emulator.cursorX),
						paneY + atlas.glyphHeight*float32(emulator.cursorY),
					}

					drawRect(renderer, cameraX, cameraY, position, glyphSize,
						color.RGBA{255, 255, 255, 255})
					drawGlyph(renderer, &atlas, cameraX, cameraY, r, position,
						color.RGBA{0, 0, 0, 255})
				}
			}

			paneY += atlas.glyphHeight * float32(emulator.usedHeight+1)
		}

		borderColor := getPaneBorderColor(len(panes), focusedPaneIndex)
		drawBorderedRect(renderer, cameraX, cameraY,
			Vector2{0, paneY}, Vector2{paneWidth, atlas.glyphHeight},
			paneBorderWidth, borderColor, color.RGBA{0, 0, 0, 255})

		if errorFlashTimer > 0 {
			errorColor := color.RGBA{255, 0, 0, uint8(errorFlashTimer * 255)}
			drawRect(renderer, cameraX, cameraY,
				Vector2{0, paneY}, Vector2{paneWidth, atlas.glyphHeight}, errorColor)
		}

		if len(command) > 0 {
			drawText(renderer, &atlas, cameraX, cameraY, command,
				Vector2{0, paneY}, color.RGBA{255, 255, 255, 255})
		}

		if len(tokenizedCommand.missingTrailingRunes) > 0 {
			drawText(renderer, &atlas, cameraX, cameraY,
				tokenizedCommand.missingTrailingRunes,
				Vector2{atlas.glyphWidth * float32(len(command)), paneY},
				color.RGBA{255, 255, 255, 255})
		}

		if len(panes) == focusedPaneIndex {
			position := Vector2{atlas.glyphWidth * float32(len(command)), paneY}

			drawRect(renderer, cameraX, cameraY, position, glyphSize,
				color.RGBA{255, 255, 255, 255})

			if len(tokenizedCommand.missingTrailingRunes) > 0 {
				drawGlyph(renderer, &atlas, cameraX, cameraY,
					tokenizedCommand.missingTrailingRunes[0], position,
					color.RGBA{0, 0, 0, 255})
			}
		}

		renderer.Present()
	}
}

func createGlyphAtlas(renderer *sdl.Renderer, font *ttf.Font) GlyphAtlas {
	const glyphsPerRow = 16
	const numGlyphs = int('~'-startGlyph) + 1

	sdlGlyphWidth, sdlGlyphHeight, _ := font.StringSize("M")
	glyphWidth, glyphHeight := int(sdlGlyphWidth), int(sdlGlyphHeight)

	atlasWidth := glyphWidth * glyphsPerRow
	atlasHeight := glyphHeight * ((numGlyphs + glyphsPerRow - 1) / glyphsPerRow)

	surface, err := sdl.CreateSurface(atlasWidth, atlasHeight, sdl.PIXELFORMAT_RGBA8888)
	if err != nil {
		panic(err)
	}
	defer surface.Destroy()

	surface.FillRect(nil, surface.MapRGBA(0, 0, 0, 0))

	atlas := GlyphAtlas{
		glyphs:      make([]sdl.FRect, numGlyphs),
		glyphWidth:  float32(glyphWidth),
		glyphHeight: float32(glyphHeight),
	}

	for i := range numGlyphs {
		r := rune(startGlyph + i)
		x := (i % glyphsPerRow) * glyphWidth
		y := (i / glyphsPerRow) * glyphHeight

		glyphSurface, err := font.RenderTextBlended(string(r), sdl.Color{R: 255, G: 255, B: 255, A: 255})
		if err != nil {
			continue
		}

		dstRect := &sdl.Rect{X: int32(x), Y: int32(y), W: glyphSurface.W, H: glyphSurface.H}
		glyphSurface.Blit(nil, surface, dstRect)
		glyphSurface.Destroy()

		atlas.glyphs[i] = sdl.FRect{
			X: float32(x),
			Y: float32(y),
			W: float32(glyphWidth),
			H: float32(glyphHeight),
		}
	}

	atlas.texture, err = renderer.CreateTextureFromSurface(surface)
	if err != nil {
		panic(err)
	}

	atlas.texture.SetBlendMode(sdl.BLENDMODE_BLEND)
	atlas.texture.SetScaleMode(sdl.SCALEMODE_LINEAR)

	return atlas
}

func handleKeyPress(key sdl.Keycode, focusedPaneIndex *int, panes *[]*pane,
	command *[]rune, tokenizedCommand *tokenizeResult, errorFlashTimer *float32,
	homeDir string, ptyInputBuffer [4]byte) {

	modState := sdl.GetModState()
	cmdPressed := (modState & sdl.KMOD_GUI) != 0

	if cmdPressed {
		switch key {
		case sdl.K_UP:
			*focusedPaneIndex = max(*focusedPaneIndex-1, 0)
		case sdl.K_DOWN:
			*focusedPaneIndex = min(*focusedPaneIndex+1, len(*panes))
		case sdl.K_W:
			if *focusedPaneIndex < len(*panes) {
				*panes = slices.Delete(*panes, *focusedPaneIndex, *focusedPaneIndex+1)
			}
		}
		return
	}

	if *focusedPaneIndex >= len(*panes) {
		switch key {
		case sdl.K_BACKSPACE:
			if len(*command) > 0 {
				*command = (*command)[:len(*command)-1]
			}
		case sdl.K_RETURN:
			tokenize(*command, tokenizedCommand)
			if runCommand(tokenizedCommand, panes, focusedPaneIndex, homeDir) {
				*command = (*command)[:0]
				tokenize(*command, tokenizedCommand)
			} else {
				*errorFlashTimer = 1
			}
		}
		tokenize(*command, tokenizedCommand)
	} else {
		pane := (*panes)[*focusedPaneIndex]
		switch key {
		case sdl.K_BACKSPACE:
			writeRuneToPty(&pane.pty, ptyInputBuffer, '\x7f')
		case sdl.K_TAB:
			writeRuneToPty(&pane.pty, ptyInputBuffer, '\t')
		case sdl.K_RETURN:
			writeRuneToPty(&pane.pty, ptyInputBuffer, '\r')
		case sdl.K_ESCAPE:
			writeRuneToPty(&pane.pty, ptyInputBuffer, '\x1b')
		}
	}
}

func drawRect(renderer *sdl.Renderer, cameraX, cameraY float32, pos, size Vector2, c color.RGBA) {
	x := pos.X - cameraX
	y := pos.Y - cameraY
	w := size.X
	h := size.Y

	renderer.SetDrawColor(c.R, c.G, c.B, c.A)
	renderer.RenderFillRect(&sdl.FRect{X: x, Y: y, W: w, H: h})
}

func drawGlyph(renderer *sdl.Renderer, atlas *GlyphAtlas, cameraX, cameraY float32,
	r rune, pos Vector2, c color.RGBA) {

	index := int(r - startGlyph)

	if index < 0 || index > len(atlas.glyphs) {
		return
	}

	srcRect := &atlas.glyphs[index]

	x := pos.X - cameraX
	y := pos.Y - cameraY

	atlas.texture.SetColorMod(c.R, c.G, c.B)
	atlas.texture.SetAlphaMod(c.A)

	dstRect := &sdl.FRect{X: x, Y: y, W: atlas.glyphWidth, H: atlas.glyphHeight}
	renderer.RenderTexture(atlas.texture, srcRect, dstRect)
}

func drawText(renderer *sdl.Renderer, atlas *GlyphAtlas, cameraX, cameraY float32,
	text []rune, pos Vector2, c color.RGBA) {

	for i, r := range text {
		glyphPos := Vector2{pos.X + atlas.glyphWidth*float32(i), pos.Y}
		drawGlyph(renderer, atlas, cameraX, cameraY, r, glyphPos, c)
	}
}

func drawBorderedRect(renderer *sdl.Renderer, cameraX, cameraY float32,
	position, size Vector2, borderWidth float32, borderColor, backgroundColor color.RGBA) {

	borderOffset := Vector2{borderWidth, borderWidth}
	borderPosition := Vector2{position.X - borderOffset.X, position.Y - borderOffset.Y}
	borderSize := Vector2{size.X + borderOffset.X*2, size.Y + borderOffset.Y*2}

	drawRect(renderer, cameraX, cameraY, borderPosition, borderSize, borderColor)
	drawRect(renderer, cameraX, cameraY, position, size, backgroundColor)
}

func getPaneBorderColor(index, focusedIndex int) color.RGBA {
	if index == focusedIndex {
		return color.RGBA{135, 206, 235, 255}
	} else {
		return color.RGBA{211, 211, 211, 255}
	}
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

func writeRuneToPty(pty *pty, buffer [4]byte, r rune) {
	size := utf8.EncodeRune(buffer[:], r)
	pty.write(buffer[:size])
}

func terminalColorToColor(c uint32) color.RGBA {
	switch c {
	case Background:
		return color.RGBA{0, 0, 0, 255}
	case Foreground:
		return color.RGBA{255, 255, 255, 255}
	case Red:
		return color.RGBA{255, 0, 0, 255}
	case Green:
		return color.RGBA{0, 255, 0, 255}
	case Yellow:
		return color.RGBA{255, 255, 0, 255}
	case Blue:
		return color.RGBA{0, 0, 255, 255}
	case Magenta:
		return color.RGBA{255, 0, 255, 255}
	case Cyan:
		return color.RGBA{135, 206, 235, 255}
	case BrightBackground:
		return brightenColor(color.RGBA{0, 0, 0, 255})
	case BrightForeground:
		return brightenColor(color.RGBA{255, 255, 255, 255})
	case BrightRed:
		return brightenColor(color.RGBA{255, 0, 0, 255})
	case BrightGreen:
		return brightenColor(color.RGBA{0, 255, 0, 255})
	case BrightYellow:
		return brightenColor(color.RGBA{255, 255, 0, 255})
	case BrightBlue:
		return brightenColor(color.RGBA{0, 0, 255, 255})
	case BrightMagenta:
		return brightenColor(color.RGBA{255, 0, 255, 255})
	case BrightCyan:
		return brightenColor(color.RGBA{135, 206, 235, 255})
	default:
		r := (c >> 16) & 0xFF
		g := (c >> 8) & 0xFF
		b := c & 0xFF

		return color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	}
}

func brightenColor(c color.RGBA) color.RGBA {
	r := brightenColorComponent(c.R)
	g := brightenColorComponent(c.G)
	b := brightenColorComponent(c.B)

	return color.RGBA{r, g, b, c.A}
}

func brightenColorComponent(x uint8) uint8 {
	return uint8((int(x)*2 + 255) / 3)
}

func lerp(a, b, t float32) float32 {
	return a + (b-a)*t
}

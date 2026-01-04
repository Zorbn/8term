package main

import (
	_ "embed"
	"io"
	"log"
	"os"
	"os/exec"
	"unicode/utf8"
	"unsafe"

	"github.com/creack/pty"
	rl "github.com/gen2brain/raylib-go/raylib"
)

//go:embed Inconsolata-Regular.ttf
var fontData []byte

//go:embed sdf.fs
var fragmentShaderCode string

func main() {
	log.Println("Hello, world!")

	rl.SetConfigFlags(rl.FlagWindowHighdpi | rl.FlagMsaa4xHint | rl.FlagVsyncHint | rl.FlagWindowResizable)

	const screenWidth, screenHeight = 800, 450
	rl.InitWindow(screenWidth, screenHeight, "raylib [text] example - Sdf fonts")
	defer rl.CloseWindow()

	const msg = "Signed Distance Fields"

	fontResolution := rl.GetWindowScaleDPI().X
	fontSize := int32(16.0 * fontResolution)

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

	scaledFontSize := float32(fontSize) / fontResolution
	glyphSize := rl.MeasureTextEx(font, "M", scaledFontSize, 0)

	var panes []*Pane

	for !rl.WindowShouldClose() {
		// Update:
		if rl.IsKeyPressed(rl.KeyP) {
			pane := newPane("git", "status")
			pane.run()

			panes = append(panes, &pane)
		}

		// Draw:
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		rl.BeginShaderMode(shader)

		paneY := 0

		for _, pane := range panes {
			for y := range pane.usedHeight {
				lineStartIndex := y * paneCols
				lineEndIndex := lineStartIndex + paneCols
				lineY := glyphSize.Y * float32(y+paneY)

				rl.DrawTextCodepoints(font, pane.grid[lineStartIndex:lineEndIndex], rl.NewVector2(0.0, lineY), scaledFontSize, 0.0, rl.Black)
			}

			paneY += pane.usedHeight
		}

		rl.EndShaderMode()

		rl.EndDrawing()

	}
}

const paneRows int = 24
const paneCols int = 80

type Pane struct {
	pty              Pty
	buffer           []byte
	grid             []rune
	usedHeight       int
	cursorX, cursorY int
}

func newPane(name string, arg ...string) Pane {
	pty := newPty(name, arg...)

	buffer := make([]byte, 4096)
	grid := make([]rune, paneRows*paneCols)

	for i := range len(grid) {
		grid[i] = ' '
	}

	usedHeight := 1
	cursorX, cursorY := 0, 0

	return Pane{
		pty,
		buffer,
		grid,
		usedHeight,
		cursorX,
		cursorY,
	}
}

func (p *Pane) run() {
	go func() {
		for {
			outputLen, err := p.pty.read(p.buffer)

			if err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			}

			output := p.buffer[:outputLen]

			p.handleOutput(output)
		}

		p.pty.tty.Close()
	}()
}

func (p *Pane) handleOutput(output []byte) {
	for i := 0; i < len(output); {
		r, size := utf8.DecodeRune(output[i:])

		if r == utf8.RuneError {
			r = '?'
		}

		i += size

		p.writeRune(r)
	}
}

func (p *Pane) writeRune(r rune) {
	if r == '\n' {
		p.newlineCursor()
		return
	}

	if p.cursorX >= paneCols {
		p.cursorX = 0
		p.cursorY++
	}

	if p.cursorY >= paneRows {
		p.scrollContentUp()
		p.cursorY--
	}

	p.grid[p.cursorY*paneCols+p.cursorX] = r
	p.cursorX++
	p.usedHeight = max(p.usedHeight, p.cursorY+1)
}

func (p *Pane) newlineCursor() {
	p.cursorX = 0
	p.cursorY++

	if p.cursorY >= paneRows {
		p.scrollContentUp()
		p.cursorY--
	}
}

func (p *Pane) scrollContentUp() {
	copy(p.grid[0:], p.grid[paneCols:])
}

type Pty struct {
	cmd *exec.Cmd
	tty *os.File
}

func newPty(name string, arg ...string) Pty {
	cmd := exec.Command(name, arg...)
	tty, err := pty.Start(cmd)

	pty.Setsize(tty, &pty.Winsize{
		Rows: uint16(paneRows),
		Cols: uint16(paneCols),
	})

	if err != nil {
		panic(err)
	}

	return Pty{
		cmd,
		tty,
	}
}

func (p *Pty) write(input []byte) {
	p.tty.Write(input)
}

func (p *Pty) read(output []byte) (int, error) {
	return p.tty.Read(output)
}

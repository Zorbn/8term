package main

import (
	"image/color"
	"math"
	"unicode"

	"github.com/Zyko0/go-sdl3/sdl"
)

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

	if unicode.IsSpace(r) {
		return
	}

	if _, ok := atlas.glyphs[r]; !ok {
		atlas.addGlyph(r)
	}

	srcRect, ok := atlas.glyphs[r]
	if !ok {
		if _, ok := atlas.glyphs['?']; !ok {
			atlas.addGlyph('?')
		}
		srcRect, ok = atlas.glyphs['?']
		if !ok {
			return
		}
	}

	x := pos.X - cameraX
	y := pos.Y - cameraY

	atlas.texture.SetColorMod(c.R, c.G, c.B)
	atlas.texture.SetAlphaMod(c.A)

	dstRect := &sdl.FRect{X: x, Y: y, W: atlas.glyphWidth, H: atlas.glyphHeight}
	renderer.RenderTexture(atlas.texture, &srcRect, dstRect)
}

func drawString(renderer *sdl.Renderer, atlas *GlyphAtlas, cameraX, cameraY float32,
	text string, pos Vector2, c color.RGBA) float32 {

	for i, r := range text {
		glyphPos := Vector2{pos.X + atlas.glyphWidth*float32(i), pos.Y}
		drawGlyph(renderer, atlas, cameraX, cameraY, r, glyphPos, c)
	}

	return atlas.glyphWidth * float32(len(text))
}

func drawText(renderer *sdl.Renderer, atlas *GlyphAtlas, cameraX, cameraY float32,
	text []rune, pos Vector2, c color.RGBA) float32 {

	for i, r := range text {
		glyphPos := Vector2{pos.X + atlas.glyphWidth*float32(i), pos.Y}
		drawGlyph(renderer, atlas, cameraX, cameraY, r, glyphPos, c)
	}

	return atlas.glyphWidth * float32(len(text))
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

func getPaneBorderWidth(index, focusedIndex int, paneBorderWidth, time float32) float32 {
	if index == focusedIndex {
		scaledTime := float64(time * 5)
		scale := (math.Sin(scaledTime) + 2) / 2

		return paneBorderWidth * float32(scale)
	} else {
		return paneBorderWidth
	}
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

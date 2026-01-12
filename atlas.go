package main

import (
	_ "embed"

	"github.com/Zyko0/go-sdl3/sdl"
	"github.com/Zyko0/go-sdl3/ttf"
)

//go:embed Inconsolata-Regular.ttf
var fontData []byte

type GlyphAtlas struct {
	renderer    *sdl.Renderer
	font        *ttf.Font
	texture     *sdl.Texture
	glyphs      map[rune]sdl.FRect
	cursorX     int32
	cursorY     int32
	rowHeight   int32
	size        int32
	glyphWidth  float32
	glyphHeight float32
}

func loadGlyphAtlas(renderer *sdl.Renderer, dpi float32) (*GlyphAtlas, error) {
	fontSize := 14 * dpi
	rwops, err := sdl.IOFromConstMem(fontData)
	if err != nil {
		return nil, err
	}

	font, err := ttf.OpenFontIO(rwops, true, fontSize)
	if err != nil {
		return nil, err
	}

	return newGlyphAtlas(renderer, font), nil
}

func newGlyphAtlas(renderer *sdl.Renderer, font *ttf.Font) *GlyphAtlas {
	sdlGlyphWidth, sdlGlyphHeight, _ := font.StringSize("M")
	glyphWidth, glyphHeight := int(sdlGlyphWidth), int(sdlGlyphHeight)

	const textureSize = 2048

	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, textureSize, textureSize)
	if err != nil {
		panic(err)
	}

	texture.SetBlendMode(sdl.BLENDMODE_BLEND)
	texture.SetScaleMode(sdl.SCALEMODE_LINEAR)

	renderer.SetRenderTarget(texture)
	renderer.SetDrawColor(0, 0, 0, 0)
	renderer.Clear()
	renderer.SetRenderTarget(nil)

	return &GlyphAtlas{
		renderer:    renderer,
		font:        font,
		texture:     texture,
		glyphs:      make(map[rune]sdl.FRect),
		size:        textureSize,
		glyphWidth:  float32(glyphWidth),
		glyphHeight: float32(glyphHeight),
		rowHeight:   int32(glyphHeight),
	}
}

func (atlas *GlyphAtlas) addGlyph(r rune) {
	if _, ok := atlas.glyphs[r]; ok {
		return
	}

	surface, err := atlas.font.RenderTextBlended(string(r), sdl.Color{R: 255, G: 255, B: 255, A: 255})
	if err != nil {
		return
	}
	defer surface.Destroy()

	if atlas.cursorX+surface.W > atlas.size {
		atlas.cursorX = 0
		atlas.cursorY += atlas.rowHeight + 1
		atlas.rowHeight = surface.H
	}

	if atlas.cursorY+surface.H > atlas.size {
		return
	}

	if surface.H > atlas.rowHeight {
		atlas.rowHeight = surface.H
	}

	tmpTexture, err := atlas.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return
	}
	defer tmpTexture.Destroy()

	tmpTexture.SetBlendMode(sdl.BLENDMODE_NONE)

	dstRect := &sdl.FRect{
		X: float32(atlas.cursorX),
		Y: float32(atlas.cursorY),
		W: float32(surface.W),
		H: float32(surface.H),
	}

	atlas.renderer.SetRenderTarget(atlas.texture)
	atlas.renderer.RenderTexture(tmpTexture, nil, dstRect)
	atlas.renderer.SetRenderTarget(nil)

	atlas.glyphs[r] = *dstRect
	atlas.cursorX += surface.W + 1
}

func (atlas *GlyphAtlas) Destroy() {
	atlas.texture.Destroy()
	atlas.font.Close()
}

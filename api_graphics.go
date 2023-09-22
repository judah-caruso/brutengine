package main

import (
	"context"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/tetratelabs/wazero/api"
)

type gfx struct {
	SetTargetSize func(w, h int32) `wasm:"GfxSetTargetSize"`

	Clear       func(c Color)                                         `wasm:"GfxClear" wrapper:"wasmClear"`
	Texture     func(tex Texture, x, y float32)                       `wasm:"GfxTexture"`
	TextureEx   func(tex Texture, x, y, rot, sx, sy float32, c Color) `wasm:"GfxTextureEx" wrapper:"wasmTextureEx"`
	Circle      func(x, y, rad float32, c Color, line bool)           `wasm:"GfxCircle" wrapper:"wasmCircle"`
	Rectangle   func(x, y, w, h float32, c Color, line bool)          `wasm:"GfxRectangle" wrapper:"wasmRectangle"`
	DefaultText func(s string, x, y float32)                          `wasm:"GfxDefaultText" wrapper:"wasmDefaultText"`
}

type Color struct {
	R, G, B, A float32
}

func (c Color) RGBA() (r, g, b, a uint32) {
	r = uint32(c.R * 255)
	r |= r << 8

	g = uint32(c.G * 255)
	g |= g << 8

	b = uint32(c.B * 255)
	b |= b << 8

	a = uint32(c.A * 255)
	a |= a << 8

	return
}

func GfxSetTargetSize(w, h int32) {
	B.ResizeRenderTarget(int(w), int(h))
}

func GfxClear(c Color) {
	B.renderTarget.Fill(c)
}

func GfxTexture(tex Texture, x, y float32) {
	GfxTextureEx(tex, x, y, 0, 1, 1, Color{1, 1, 1, 1})
}

func GfxTextureEx(tex Texture, x, y, rot, sx, sy float32, c Color) {
	handle, ok := B.GetTextureById(tex)
	if !ok {
		return
	}

	o := &B.renderOpts
	o.GeoM.Reset()
	o.ColorScale.Reset()

	if rot != 0 {
		bounds := handle.Bounds().Size()
		o.GeoM.Translate(-float64(bounds.X)/2, -float64(bounds.Y)/2)
		o.GeoM.Rotate(float64(rot))
	}

	o.GeoM.Translate(float64(x), float64(y))
	o.GeoM.Scale(float64(sx), float64(sy))
	o.ColorScale.Scale(c.R, c.G, c.B, 1)
	o.ColorScale.ScaleAlpha(c.A)

	B.renderTarget.DrawImage(handle, o)
}

func GfxRectangle(x, y, w, h float32, c Color, line bool) {
	if line {
		vector.StrokeRect(B.renderTarget, x, y, w, h, 1, c, false)
	} else {
		vector.DrawFilledRect(B.renderTarget, x, y, w, h, c, false)
	}
}

func GfxCircle(x, y, rad float32, c Color, line bool) {
	if line {
		vector.StrokeCircle(B.renderTarget, x, y, rad, 1, c, false)
	} else {
		vector.DrawFilledCircle(B.renderTarget, x, y, rad, c, false)
	}
}

func GfxDefaultText(s string, x, y float32) {
	ebitenutil.DebugPrintAt(B.renderTarget, s, int(x), int(y))
}

// Wasm wrappers

func wasmClear(r, g, b, a float32) {
	GfxClear(Color{R: r, G: g, B: b, A: a})
}

func wasmRectangle(x, y, w, h float32, r, g, b, a float32, line int32) {
	GfxRectangle(x, y, w, h, Color{R: r, G: g, B: b, A: a}, line == 1)
}

func wasmTextureEx(tex Texture, x, y, rot, sx, sy float32, r, g, b, a float32) {
	GfxTextureEx(tex, x, y, rot, sx, sy, Color{R: r, G: g, B: b, A: a})
}

func wasmCircle(x, y, rad float32, r, g, b, a float32, line int32) {
	GfxCircle(x, y, rad, Color{R: r, G: g, B: b, A: a}, line == 1)
}

func wasmDefaultText(_ context.Context, m api.Module, offset, count uint32, x, y float32) {
	str := readWasmString(m.Memory(), offset, count)
	GfxDefaultText(str, x, y)
}

package engine

import (
	"errors"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type IGraphics interface {
	SetTargetSize(width, height int32)
	Clear(c Color)
	Texture(tex Texture, x, y float32)
	TextureEx(tex Texture, x, y, rot, sx, sy float32, c Color)
	Rectangle(x, y, w, h float32, c Color, line bool)
	Circle(x, y, rad float32, c Color, line bool)
	Text(str string, x, y float32)
}

type Graphics struct {
	Target                    *ebiten.Image
	TargetWidth, TargetHeight int

	opts ebiten.DrawImageOptions
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

func (g *Graphics) Setup() error {
	g.TargetWidth = 960
	g.TargetHeight = 540

	g.Target = ebiten.NewImage(g.TargetWidth, g.TargetHeight)
	if g.Target == nil {
		return errors.New("graphics - unable to create render target ")
	}

	return nil
}

func (g *Graphics) Present(screen *ebiten.Image) {
	g.opts.GeoM.Reset()
	g.opts.ColorScale.Reset()
	g.opts.Blend.BlendOperationAlpha = ebiten.BlendOperationAdd
	screen.DrawImage(g.Target, &g.opts)
}

func (g *Graphics) SetTargetSize(w, h int32) {
	LogDebug("graphics - resizing render target")

	if g.Target != nil {
		g.Target.Dispose()
	}

	g.TargetWidth = int(w)
	g.TargetHeight = int(h)

	g.Target = ebiten.NewImage(g.TargetWidth, g.TargetHeight)
	if g.Target == nil {
		LogError("graphics - unable to resize render target to %d, %d", w, h)
	}
}

func (g *Graphics) Clear(c Color) {
	g.Target.Fill(c)
}

func (g *Graphics) Texture(tex Texture, x, y float32) {
	g.TextureEx(tex, x, y, 0, 1, 1, Color{R: 1, G: 1, B: 1, A: 1})
}

func (g *Graphics) TextureEx(tex Texture, x, y, rot, sx, sy float32, c Color) {
	handle, ok := brut.Asset.GetTextureData(tex)
	if !ok {
		return
	}

	o := &g.opts
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

	g.Target.DrawImage(handle, o)
}

func (g *Graphics) Text(s string, x, y float32) {
	ebitenutil.DebugPrintAt(g.Target, s, int(x), int(y))
}

func (g *Graphics) Rectangle(x, y, w, h float32, c Color, line bool) {
	if line {
		vector.StrokeRect(g.Target, x, y, w, h, 1, c, false)
	} else {
		vector.DrawFilledRect(g.Target, x, y, w, h, c, false)
	}
}

func (g *Graphics) Circle(x, y, rad float32, c Color, line bool) {
	if line {
		vector.StrokeCircle(g.Target, x, y, rad, 1, c, false)
	} else {
		vector.DrawFilledCircle(g.Target, x, y, rad, c, false)
	}
}

// Wasm api

func (*Graphics) Namespace() string {
	return "Graphics"
}

// Used to ensure Graphics implements IGraphics correctly
var _ IGraphics = (*Graphics)(nil)

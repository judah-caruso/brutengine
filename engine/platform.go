package engine

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// IPlatform describes the public go api.
type IPlatform interface {
	SetTitle(title string)
	SetScreenSize(width, height int32)
	Log(msg string)
	Fps() float32
	Tps() float32
	Exit()
}

type Platform struct {
	ExitRequested             bool
	ScreenWidth, ScreenHeight int
}

func (p *Platform) Setup() error {
	p.ScreenWidth = 960
	p.ScreenHeight = 540
	p.ExitRequested = false

	ebiten.SetWindowSize(p.ScreenWidth, p.ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	return nil
}

func (p *Platform) SetTitle(title string) {
	ebiten.SetWindowTitle(title)
}

func (p *Platform) SetScreenSize(w, h int32) {
	p.ScreenWidth = int(w)
	p.ScreenHeight = int(h)

	if ebiten.IsFullscreen() {
		ebiten.SetFullscreen(false)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled) // fixes issue with resizing
	ebiten.SetWindowSize(p.ScreenWidth, p.ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}

func (*Platform) Log(msg string) {
	fmt.Println(msg)
}

func (p *Platform) Exit() {
	p.ExitRequested = true
}

func (*Platform) Fps() float32 {
	return float32(ebiten.ActualFPS())
}

func (*Platform) Tps() float32 {
	return float32(ebiten.ActualTPS())
}

// Used to ensure Platform implements IPlatform correctly
var _ IPlatform = (*Platform)(nil)

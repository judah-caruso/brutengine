package main

import (
	"context"
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tetratelabs/wazero/api"
)

type platform struct {
	SetWindowSize func(w, h int32) `wasm:"PlatformSetWindowSize"`
	Log           func(s string)   `wasm:"PlatformLog" wrapper:"wasmLog"`
	Exit          func()           `wasm:"PlatformExit"`
	Fps           func() float32   `wasm:"PlatformFps"`
	Tps           func() float32   `wasm:"PlatformTps"`
}

func PlatformSetWindowSize(w, h int32) {
	B.screenWidth = int(w)
	B.screenHeight = int(h)

	if ebiten.IsFullscreen() {
		ebiten.SetFullscreen(false)
	}

	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled) // fixes issue with resizing
	ebiten.SetWindowSize(B.screenWidth, B.screenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
}

func PlatformLog(s string) {
	fmt.Println(s)
}

func PlatformExit() {
	B.shouldExit = true
}

func PlatformFps() float32 {
	return float32(ebiten.ActualFPS())
}

func PlatformTps() float32 {
	return float32(ebiten.ActualTPS())
}

// Wasm wrappers

func wasmLog(_ context.Context, m api.Module, offset, count uint32) {
	str := readWasmString(m.Memory(), offset, count)
	PlatformLog(str)
}

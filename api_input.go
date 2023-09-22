package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type input struct {
	Up      func(i Input) bool `wasm:"InputUp" wrapper:"wasmUp"`
	Down    func(i Input) bool `wasm:"InputDown" wrapper:"wasmDown"`
	Pressed func(i Input) bool `wasm:"InputPressed" wrapper:"wasmPressed"`
	CursorX func() float32     `wasm:"InputCursorX"`
	CursorY func() float32     `wasm:"InputCursorY"`
}

type Input int32

const (
	InputNone Input = iota

	_inputKeyboardStart
	InputEscape
	InputEnter
	InputSpace
	InputBackspace
	_inputKeyboardEnd

	_inputMouseStart
	InputMouseLeft
	InputMouseMiddle
	InputMouseRight
	_inputMouseEnd
)

func (i Input) IsKeyboard() bool {
	return i >= _inputKeyboardStart && i <= _inputKeyboardEnd
}

func (i Input) IsMouse() bool {
	return i >= _inputMouseStart && i <= _inputMouseEnd
}

func InputUp(i Input) (up bool) {
	c, ok := inputMap[i]
	if !ok {
		return true
	}

	switch {
	case i.IsKeyboard():
		up = !ebiten.IsKeyPressed(ebiten.Key(c))
	case i.IsMouse():
		up = !ebiten.IsMouseButtonPressed(ebiten.MouseButton(c))
	}

	return
}

func InputDown(i Input) (down bool) {
	c, ok := inputMap[i]
	if !ok {
		return false
	}

	switch {
	case i.IsKeyboard():
		down = ebiten.IsKeyPressed(ebiten.Key(c))
	case i.IsMouse():
		down = ebiten.IsMouseButtonPressed(ebiten.MouseButton(c))
	}

	return
}

func InputPressed(i Input) (pressed bool) {
	c, ok := inputMap[i]
	if !ok {
		return false
	}

	switch {
	case i.IsKeyboard():
		pressed = inpututil.IsKeyJustPressed(ebiten.Key(c))
	case i.IsMouse():
		pressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButton(c))
	}

	return
}

func InputCursorX() float32 {
	return B.cursorX
}

func InputCursorY() float32 {
	return B.cursorY
}

// Wasm wrappers

func wasmUp(i Input) int32 {
	if InputUp(i) {
		return 1
	}

	return 0
}

func wasmDown(i Input) int32 {
	if InputDown(i) {
		return 1
	}

	return 0
}

func wasmPressed(i Input) int32 {
	if InputPressed(i) {
		return 1
	}

	return 0
}

var inputMap = map[Input]int{
	InputNone: 0,

	InputEscape:    int(ebiten.KeyEscape),
	InputEnter:     int(ebiten.KeyEnter),
	InputSpace:     int(ebiten.KeySpace),
	InputBackspace: int(ebiten.KeyBackspace),

	InputMouseLeft:   int(ebiten.MouseButtonLeft),
	InputMouseMiddle: int(ebiten.MouseButtonMiddle),
	InputMouseRight:  int(ebiten.MouseButtonRight),
}

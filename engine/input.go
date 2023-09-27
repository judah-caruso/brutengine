package engine

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type (
	Input struct {
		cursorX, cursorY     float32
		thisFrame, lastFrame [_inputMax + 1]inputState
	}
	IInput interface {
		Pressed(InputEvent) bool
		Up(InputEvent) bool
		Down(InputEvent) bool
		CursorX() float32
		CursorY() float32
	}
)

type InputEvent = uint32

const (
	InputNone InputEvent = iota

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

	_inputMax
)

type inputState uint32

const (
	stateDown inputState = 1 << iota
	stateControl
	stateShift
	stateAlt
)

func (i *Input) Setup() error {
	return nil
}

func (i *Input) Pressed(e InputEvent) bool {
	last := i.lastFrame[e]&stateDown != 0
	this := i.thisFrame[e]&stateDown != 0
	return last && !this
}

func (i *Input) Up(e InputEvent) bool {
	last := i.lastFrame[e]&stateDown != 0
	this := i.thisFrame[e]&stateDown != 0
	return !last && !this
}

func (i *Input) Down(e InputEvent) bool {
	last := i.lastFrame[e]&stateDown != 0
	this := i.thisFrame[e]&stateDown != 0
	return last && this
}

func (i *Input) CursorX() float32 {
	return i.cursorX
}

func (i *Input) CursorY() float32 {
	return i.cursorY
}

func (i *Input) Update() {
	// Transfer/reset state
	copy(i.lastFrame[:], i.thisFrame[:])
	clear(i.thisFrame[:])

	cx, cy := ebiten.CursorPosition()
	i.cursorX = float32(cx)
	i.cursorY = float32(cy)

	var modState inputState

	if inpututil.KeyPressDuration(ebiten.KeyControl) >= 1 {
		modState |= stateControl
	}

	if inpututil.KeyPressDuration(ebiten.KeyShift) >= 1 {
		modState |= stateShift
	}

	if inpututil.KeyPressDuration(ebiten.KeyAlt) >= 1 {
		modState |= stateAlt
	}

	for code, e := range eventMap {
		var state inputState

		switch {
		case e >= _inputKeyboardStart && e <= _inputKeyboardEnd:
			if inpututil.KeyPressDuration(ebiten.Key(code)) >= 1 {
				state = stateDown
			}
		case e >= _inputMouseStart && e <= _inputMouseEnd:
			if inpututil.MouseButtonPressDuration(ebiten.MouseButton(code)) >= 1 {
				state = stateDown
			}
		}

		i.thisFrame[e] = state | modState
	}
}

var eventMap = map[int]InputEvent{
	int(ebiten.KeyEscape):    InputEscape,
	int(ebiten.KeyBackspace): InputBackspace,
	int(ebiten.KeySpace):     InputSpace,

	int(ebiten.MouseButtonLeft):   InputMouseLeft,
	int(ebiten.MouseButtonMiddle): InputMouseMiddle,
	int(ebiten.MouseButtonRight):  InputMouseRight,
}

var _ IInput = (*Input)(nil)

package main

// Platform

//go:export PlatformLog
func PlatformLog(string)

//go:export PlatformSetScreenSize
func PlatformSetScreenSize(w, h int32)

//go:export PlatformExit
func PlatformExit()

//go:export PlatformFps
func PlatformFps() float32

//go:export PlatformTps
func PlatformTps() float32

// Input

type (
	Input int32
)

const (
	InputNone Input = 0

	InputEscape    Input = 2
	InputEnter     Input = 3
	InputSpace     Input = 4
	InputBackspace Input = 5

	InputMouseLeft   Input = 8
	InputMouseMiddle Input = 9
	InputMouseRight  Input = 10
)

//go:export InputUp
func InputUp(Input) bool

//go:export InputDown
func InputDown(Input) bool

//go:export InputPressed
func InputPressed(Input) bool

//go:export InputCursorX
func InputCursorX() float32

//go:export InputCursorY
func InputCursorY() float32

// Graphics

//go:export GraphicsSetTargetSize
func GraphicsSetTargetSize(int32, int32)

//go:export GraphicsClear
func GraphicsClear(r, g, b, a float32)

//go:export GraphicsTexture
func GraphicsTexture(tex Texture, x, y float32)

//go:export GraphicsTextureEx
func GraphicsTextureEx(tex Texture, x, y, rot, sx, sy float32, r, g, b, a float32)

//go:export GraphicsRectangle
func GraphicsRectangle(x, y, w, h float32, r, g, b, a float32, line bool)

//go:export GraphicsCircle
func GraphicsCircle(x, y, rad float32, r, g, b, a float32, line bool)

//go:export GraphicsText
func GraphicsText(s string, x, y float32)

// Asset

type (
	Texture uint32
)

//go:export AssetLoadTexture
func AssetLoadTexture(string) Texture

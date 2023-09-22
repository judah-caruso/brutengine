package main

// Config

//go:export ConfigSetFlags
func ConfigSetFlags(ConfigFlag)

type ConfigFlag uint32

const (
	ConfigHotReloading ConfigFlag = 1 << iota
	ConfigEngineLogging
)

// Platform

//go:export PlatformLog
func PlatformLog(string)

//go:export PlatformSetWindowSize
func PlatformSetWindowSize(w, h int32)

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

//go:export GfxSetTargetSize
func GfxSetTargetSize(int32, int32)

//go:export GfxClear
func GfxClear(r, g, b, a float32)

//go:export GfxTexture
func GfxTexture(tex Texture, x, y float32)

//go:export GfxTextureEx
func GfxTextureEx(tex Texture, x, y, rot, sx, sy float32, r, g, b, a float32)

//go:export GfxRectangle
func GfxRectangle(x, y, w, h float32, r, g, b, a float32, line bool)

//go:export GfxCircle
func GfxCircle(x, y, rad float32, r, g, b, a float32, line bool)

//go:export GfxDefaultText
func GfxDefaultText(s string, x, y float32)

// Asset

type (
	Texture uint32
)

//go:export AssetLoadImage
func AssetLoadImage(string) Texture

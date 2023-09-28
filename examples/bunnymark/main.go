package main

import (
	"math"
	"math/rand"
	"strconv"
	_ "unsafe"

	brut "bunnymark/brutengine_go"
)

const (
	GRAVITY               = 0.0981
	SCREEN_WIDTH  float32 = 960
	SCREEN_HEIGHT float32 = 540
	SPRITE_SIZE   float32 = 27
)

var (
	defaultEntities   = 1000
	entitiesPerSecond = 100
	entityTex         brut.Texture

	entities = make([]Entity, defaultEntities)
	colors   = []brut.Color{
		brut.Color{1, 0.25, 0.25, 1},
		brut.Color{0.25, 1, 0.25, 1},
		brut.Color{0.25, 0.25, 1, 1},
	}
)

type Entity struct {
	x, y   float32
	vx, vy float32
	c      brut.Color
	t      float32
}

//go:export Config
func config() {
}

//go:export Setup
func setup() {
	brut.PlatformLog("Wasm setup")

	brut.PlatformSetTitle("Gophermark")
	brut.GraphicsSetTargetSize(int32(SCREEN_WIDTH), int32(SCREEN_HEIGHT))
	brut.PlatformSetScreenSize(int32(SCREEN_WIDTH), int32(SCREEN_HEIGHT))

	entityTex = brut.AssetLoadTexture("gopher.png")
	if entityTex == 0 {
		brut.PlatformLog("unable to load bunny asset")
		brut.PlatformExit()
	}

	for i := range entities {
		e := &entities[i]
		e.x = rand.Float32() * SCREEN_WIDTH
		e.y = rand.Float32() * SCREEN_HEIGHT
		e.vx = rand.Float32() * 8
		e.vy = rand.Float32() * 8
		e.c = colors[rand.Intn(len(colors))]
	}
}

//go:export Teardown
func teardown() {
	brut.PlatformLog("Wasm teardown")
}

var timer = 1.0

//go:export Update
func update() {
	if brut.InputPressed(brut.InputEventEscape) {
		brut.PlatformExit()
	}

	timer -= 0.1
	if timer <= 0 || brut.InputDown(brut.InputEventMouseLeft) {
		x := brut.InputCursorX()
		y := brut.InputCursorY()

		ents := make([]Entity, entitiesPerSecond)
		for i := range ents {
			ents[i] = Entity{
				x:  rand.Float32() * SCREEN_WIDTH,
				y:  rand.Float32() * SCREEN_HEIGHT,
				vx: rand.Float32() * 8,
				vy: rand.Float32() * 8,
				c:  colors[rand.Intn(len(colors))],
			}

			if timer > 0 {
				ents[i].x = x
				ents[i].y = y
			}
		}

		entities = append(entities, ents...)
		timer = 1
	}

	for i := range entities {
		e := &entities[i]

		if e.t < 1 {
			e.t += 0.01
		}

		e.vy += GRAVITY
		e.x += e.vx
		e.y += e.vy

		if e.y >= SCREEN_HEIGHT-SPRITE_SIZE/2 {
			e.vy *= 0.85 / 2
			if rand.Float32() > 0.5 {
				e.vy -= rand.Float32() * 8
			}
		} else if e.y < 0 {
			e.vy = -e.vy
		}

		if e.x >= SCREEN_WIDTH-SPRITE_SIZE/2 {
			e.vx = -float32(math.Abs(float64(e.vx)))
		} else if e.y < 0 {
			e.vx = float32(math.Abs(float64(e.vx)))
		}
	}
}

//go:export Render
func render() {
	brut.GraphicsClear(.12, .12, .12, 1)

	for _, e := range entities {
		brut.GraphicsTextureEx(entityTex, e.x, e.y, 0, 1, 1, e.c.C1, e.c.C2, e.c.C3, e.c.C4*e.t)
	}

	brut.GraphicsRectangle(10, 10, 120, 60, 0, 0, 0, 0.5, false)

	fps := strconv.FormatFloat(float64(brut.PlatformFps()), 'f', 2, 32)
	brut.GraphicsText("fps: "+fps, 10, 10)

	tps := strconv.FormatFloat(float64(brut.PlatformTps()), 'f', 2, 32)
	brut.GraphicsText("tps: "+tps, 10, 24)

	brut.GraphicsText("entities: "+strconv.FormatInt(int64(len(entities)), 10), 10, 36)

	x := brut.InputCursorX()
	y := brut.InputCursorY()

	xs := strconv.FormatFloat(float64(x), 'f', 2, 32)
	ys := strconv.FormatFloat(float64(y), 'f', 2, 32)
	brut.GraphicsText("x "+xs+", y "+ys, 10, 48)
}

func main() {}

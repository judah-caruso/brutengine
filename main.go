package main

import (
	"errors"
	"github.com/fsnotify/fsnotify"
	eb "github.com/hajimehoshi/ebiten/v2"
	"log"
	"strings"
	"sync"
)

var (
	B                   BrutEngine
	defaultTargetWidth  = 800
	defaultTargetHeight = 600
)

type (
	BrutEngine struct {
		// config values
		flags ConfigFlag

		wasm *WasmRuntime
		mut  sync.Mutex

		renderTarget *eb.Image
		renderOpts   eb.DrawImageOptions

		screenWidth, screenHeight int
		renderWidth, renderHeight int
		cursorX, cursorY          float32
		shouldExit                bool
		wasmSetupCalled           bool

		loadedTextures map[Texture]textureData
	}
)

func (b *BrutEngine) Setup() error {
	w, err := NewRuntime("game.wasm")
	if err != nil {
		return err
	}

	b.wasm = w
	b.wasm.CallConfig()

	if b.flags&ConfigEngineLogging != 0 {
		AddLogLevel(LevelAll)
	} else {
		AddLogLevel(LevelError)
	}

	if b.flags&ConfigHotReloading != 0 {
		LogDebug("hot reloading is enabled")

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return err
		}

		err = watcher.Add("game.wasm")
		if err != nil {
			return errors.Join(err, watcher.Close())
		}

		go b.watchForChanges(watcher)
	}

	b.renderWidth = defaultTargetWidth
	b.renderHeight = defaultTargetHeight
	b.loadedTextures = make(map[Texture]textureData)
	return nil
}

func (b *BrutEngine) Teardown() {
	b.wasm.Close()
}

func (b *BrutEngine) watchForChanges(watcher *fsnotify.Watcher) {
	watchList := strings.Join(watcher.WatchList(), ", ")
	LogDebug("watch %s for changes", watchList)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				b.mut.Lock()

				LogDebug("reloading %q", event.Name)

				err := b.wasm.Reload()
				if err != nil {
					LogWarn("%s", err)
				}

				b.wasmSetupCalled = false
				b.mut.Unlock()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}

			LogError("watcher: %s", err)
		}
	}
}

func (b *BrutEngine) GetTextureById(tex Texture) (*eb.Image, bool) {
	data, ok := b.loadedTextures[tex]
	if ok {
		return data.handle, true
	}

	return nil, false
}

func (b *BrutEngine) GetTextureByName(name string) (Texture, bool) {
	for id, data := range b.loadedTextures {
		if data.path == name {
			return id, true
		}
	}

	return InvalidTexture, false
}

func (b *BrutEngine) Start() error {
	opts := eb.RunGameOptions{
		GraphicsLibrary:   eb.GraphicsLibraryAuto,
		InitUnfocused:     false,
		ScreenTransparent: false,
		SkipTaskbar:       false,
	}

	defer b.wasm.CallTeardown()

	if err := eb.RunGameWithOptions(b, &opts); !errors.Is(err, eb.Termination) {
		return err
	}

	return nil
}

func (b *BrutEngine) ResizeRenderTarget(w, h int) {
	LogDebug("resizing render target")

	if b.renderTarget != nil {
		b.renderTarget.Dispose()
	}

	b.renderWidth = w
	b.renderHeight = h
	b.renderTarget = eb.NewImage(w, h)
	if b.renderTarget == nil {
		LogError("unable to resize render target to %d, %d", w, h)
	}
}

func (b *BrutEngine) Update() error {
	if b.shouldExit {
		return eb.Termination
	}

	if !b.wasmSetupCalled {
		b.wasm.CallSetup()

		if b.renderTarget == nil {
			b.ResizeRenderTarget(defaultTargetWidth, defaultTargetHeight)
		}

		b.wasmSetupCalled = true
	}

	{
		cx, cy := eb.CursorPosition()
		b.cursorX = float32(cx)
		b.cursorY = float32(cy)
	}

	b.wasm.CallUpdate()
	return nil
}

func (b *BrutEngine) Draw(dest *eb.Image) {
	if b.renderTarget != nil {
		b.wasm.CallRender()

		b.renderOpts.GeoM.Reset()
		b.renderOpts.ColorScale.Reset()
		b.renderOpts.Blend.BlendOperationAlpha = eb.BlendOperationAdd
		dest.DrawImage(b.renderTarget, &b.renderOpts)
	}
}

func (b *BrutEngine) Layout(dw, dh int) (rw, rh int) {
	return b.renderWidth, b.renderHeight
}

func main() {
	err := B.Setup()
	if err != nil {
		log.Fatal(err)
	}

	defer B.Teardown()

	err = B.Start()
	if err != nil {
		log.Fatal(err)
	}
}

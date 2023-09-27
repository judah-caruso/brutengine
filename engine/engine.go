package engine

import (
	"errors"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	eb "github.com/hajimehoshi/ebiten/v2"
	"github.com/pkg/profile"
)

// B is the global engine instance
var brut BrutEngine

type BrutEngine struct {
	mut              sync.Mutex
	needsToCallSetup bool
	wasm             *WasmRuntime

	Config   Config
	Platform Platform
	Input    Input
	Asset    Asset
	Graphics Graphics
}

func Setup() error {
	p := profile.Start(profile.ProfilePath("."))
	defer p.Stop()

	// Init
	{
		w, err := NewWasmRuntime(
			"game.wasm",
			&brut.Config,
			&brut.Platform,
			&brut.Input,
			&brut.Asset,
			&brut.Graphics,
		)

		if err != nil {
			return err
		}

		err = errors.Join(err, brut.Config.Setup())
		err = errors.Join(err, brut.Platform.Setup())
		err = errors.Join(err, brut.Input.Setup())
		err = errors.Join(err, brut.Asset.Setup())
		err = errors.Join(err, brut.Graphics.Setup())
		if err != nil {
			return err
		}

		brut.wasm = w
	}

	// Configure
	{
		brut.wasm.CallConfig()

		cfg := brut.Config

		if cfg.Engine&EngineLogging != 0 {
			AddLogLevel(LevelAll)
		} else {
			AddLogLevel(LevelError)
		}

		if cfg.Engine&EngineHotReload != 0 {
			LogDebug("engine - hot reloading is enabled")

			watcher, err := fsnotify.NewWatcher()
			if err != nil {
				LogWarn("engine - unable to setup watcher: %s", err)
				goto setupEnd
			}

			err = watcher.Add("game.wasm")
			if err != nil {
				LogWarn("engine - unable to watch game.wasm: %s", err)
				goto setupEnd
			}

			go brut.watchForChanges(watcher)
		}

	setupEnd:
		// Call user setup after configuration so all setup is done before the window opens
		brut.wasm.CallSetup()
	}

	return nil
}

func Run() error {
	opts := eb.RunGameOptions{
		GraphicsLibrary:   eb.GraphicsLibraryAuto,
		InitUnfocused:     false,
		ScreenTransparent: false,
		SkipTaskbar:       false,
	}

	if err := eb.RunGameWithOptions(&brut, &opts); err != nil && !errors.Is(err, eb.Termination) {
		return err
	}

	return nil
}

func Teardown() {
	brut.wasm.Teardown()
}

func (b *BrutEngine) Update() error {
	if b.Platform.ExitRequested {
		return eb.Termination
	}

	if b.needsToCallSetup {
		b.wasm.CallSetup()
		b.needsToCallSetup = false
	}

	b.Input.Update()
	b.wasm.CallUpdate()
	return nil
}

func (b *BrutEngine) Draw(dest *eb.Image) {
	if b.Graphics.Target != nil {
		b.wasm.CallRender()
		b.Graphics.Present(dest)
	}
}

func (b *BrutEngine) Layout(dw, dh int) (rw, rh int) {
	return b.Graphics.TargetWidth, b.Graphics.TargetHeight
}

func (b *BrutEngine) watchForChanges(watcher *fsnotify.Watcher) {
	watchList := strings.Join(watcher.WatchList(), ", ")
	LogDebug("engine - watching %s for changes", watchList)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				continue
			}

			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				b.mut.Lock()

				LogDebug("engine - reloading %q", event.Name)

				needsSetup := b.Config.Engine&EngineSetupAfterReload != 0
				err := b.wasm.Reload(!needsSetup)
				if err != nil {
					LogWarn("%s", err)
				}

				b.needsToCallSetup = needsSetup
				b.mut.Unlock()
			}

		case err, ok := <-watcher.Errors:
			if !ok {
				continue
			}

			b.mut.Lock()
			LogError("engine - %s", err)
			b.mut.Unlock()
		}
	}
}

//go:generate go test -run "TestGenerateWasmApi"
package main

import (
	"context"
	"errors"
	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
	"os"
)

/*

This file describes the Wasm Api exposed by BrutEngine for games.

*/

type WasmRuntime struct {
	ctx      context.Context
	rt       wazero.Runtime
	mod      api.Module
	compiled wazero.CompiledModule
	filename string

	cbConfig,
	cbSetup,
	cbTeardown,
	cbUpdate,
	cbRender api.Function

	stack []uint64
}

func NewRuntime(filename string) (*WasmRuntime, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	rt := wazero.NewRuntime(ctx)

	// Makes importing other languages much easier
	_, err = wasi_snapshot_preview1.Instantiate(ctx, rt)
	if err != nil {
		return nil, err
	}

	b := rt.NewHostModuleBuilder("env")

	exportConfig(b)
	exportPlatform(b)
	exportInput(b)
	exportGfx(b)
	exportAsset(b)

	compiled, err := b.Compile(ctx)
	if err != nil {
		return nil, err
	}

	_, err = rt.InstantiateModule(ctx, compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, err
	}

	mod, err := rt.Instantiate(ctx, src)
	if err != nil {
		return nil, err
	}

	// Allow lower/uppercase versions of callbacks
	config := mod.ExportedFunction("config")
	if config == nil {
		config = mod.ExportedFunction("Config")
	}

	setup := mod.ExportedFunction("setup")
	if setup == nil {
		setup = mod.ExportedFunction("Setup")
	}

	teardown := mod.ExportedFunction("teardown")
	if teardown == nil {
		teardown = mod.ExportedFunction("Teardown")
	}

	update := mod.ExportedFunction("update")
	if update == nil {
		update = mod.ExportedFunction("Update")
	}

	render := mod.ExportedFunction("render")
	if render == nil {
		render = mod.ExportedFunction("Render")
	}

	return &WasmRuntime{
		ctx:      ctx,
		rt:       rt,
		mod:      mod,
		compiled: compiled,
		filename: filename,

		cbConfig:   config,
		cbSetup:    setup,
		cbTeardown: teardown,
		cbUpdate:   update,
		cbRender:   render,
		stack:      make([]uint64, 16),
	}, nil
}

func (w *WasmRuntime) Close() {
	_ = w.mod.Close(w.ctx)
	_ = w.compiled.Close(w.ctx)
	_ = w.rt.Close(w.ctx)
}

func (w *WasmRuntime) Reload() error {
	if w.compiled == nil {
		return errors.New("attempt to reload module before it has been loaded")
	}

	src, err := os.ReadFile(w.filename)
	if err != nil {
		return err
	}

	newMod, err := w.rt.Instantiate(w.ctx, src)
	if err != nil {
		return err
	}

	err = w.mod.Close(w.ctx)
	if err != nil {
		return err
	}

	w.mod = newMod

	// Reload callbacks
	w.cbConfig = w.mod.ExportedFunction("config")
	if w.cbConfig == nil {
		w.cbConfig = w.mod.ExportedFunction("Config")
	}

	w.cbSetup = w.mod.ExportedFunction("setup")
	if w.cbSetup == nil {
		w.cbSetup = w.mod.ExportedFunction("Setup")
	}

	w.cbTeardown = w.mod.ExportedFunction("teardown")
	if w.cbTeardown == nil {
		w.cbTeardown = w.mod.ExportedFunction("Teardown")
	}

	w.cbUpdate = w.mod.ExportedFunction("update")
	if w.cbUpdate == nil {
		w.cbUpdate = w.mod.ExportedFunction("Update")
	}

	w.cbRender = w.mod.ExportedFunction("render")
	if w.cbRender == nil {
		w.cbRender = w.mod.ExportedFunction("Render")
	}

	return nil
}

func (w *WasmRuntime) CallConfig() {
	if w.cbConfig == nil {
		return
	}

	w.invokeCallback(w.cbConfig)
}

func (w *WasmRuntime) CallSetup() {
	if w.cbSetup == nil {
		return
	}

	w.invokeCallback(w.cbSetup)
}

func (w *WasmRuntime) CallTeardown() {
	if w.cbTeardown == nil {
		return
	}

	w.invokeCallback(w.cbTeardown)
}

func (w *WasmRuntime) CallUpdate() {
	if w.cbUpdate == nil {
		return
	}

	w.invokeCallback(w.cbUpdate)
}

func (w *WasmRuntime) CallRender() {
	if w.cbRender == nil {
		return
	}

	w.invokeCallback(w.cbRender)
}

func (w *WasmRuntime) invokeCallback(cb api.Function) {
	clear(w.stack)

	err := cb.CallWithStack(w.ctx, w.stack)
	if err != nil {
		LogError("%s", err)
	}
}

func readWasmString(m api.Memory, offset, count uint32) string {
	buf, ok := m.Read(offset, count)
	if !ok {
		LogError("invalid memory read of %d bytes at %d", count, offset)
		return ""
	}

	return string(buf)
}

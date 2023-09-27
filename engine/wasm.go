//go:generate go run generate/generate.go
package engine

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
	"github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

/*

This file describes the Wasm Api exposed by BrutEngine for games.

*/

type (
	WasmRuntime struct {
		ctx  context.Context
		rt   wazero.Runtime
		mod  api.Module
		host wazero.HostModuleBuilder

		filename string
		compiled wazero.CompiledModule

		cbConfig,
		cbSetup,
		cbTeardown,
		cbUpdate,
		cbRender api.Function

		stack []uint64
	}
	WasmModule interface {
		Expose(*WasmRuntime)
	}
	WasmValue = uint64
	WasmType  = api.ValueType
)

var (
	WasmF32 = api.ValueTypeF32
	WasmU32 = api.ValueTypeI32
	WasmI32 = api.ValueTypeI32
)

func NewWasmRuntime(filename string, modules ...WasmModule) (*WasmRuntime, error) {
	src, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var wasm = &WasmRuntime{
		filename: filename,
		stack:    make([]uint64, 16),
		ctx:      context.Background(),
	}

	wasm.rt = wazero.NewRuntime(wasm.ctx)

	// Makes importing things much easier
	_, err = wasi_snapshot_preview1.Instantiate(wasm.ctx, wasm.rt)
	if err != nil {
		return nil, err
	}

	wasm.host = wasm.rt.NewHostModuleBuilder("env")

	// Import engine api into 'env'
	for _, mod := range modules {
		mod.Expose(wasm)
	}

	wasm.compiled, err = wasm.host.Compile(wasm.ctx)
	if err != nil {
		return nil, err
	}

	_, err = wasm.rt.InstantiateModule(wasm.ctx, wasm.compiled, wazero.NewModuleConfig())
	if err != nil {
		return nil, err
	}

	// Instantiate user's wasm module
	wasm.mod, err = wasm.rt.Instantiate(wasm.ctx, src)
	if err != nil {
		return nil, err
	}

	// Allow lower/uppercase versions of callbacks
	wasm.cbConfig = wasm.mod.ExportedFunction("config")
	if wasm.cbConfig == nil {
		wasm.cbConfig = wasm.mod.ExportedFunction("Config")
	}

	wasm.cbSetup = wasm.mod.ExportedFunction("setup")
	if wasm.cbSetup == nil {
		wasm.cbSetup = wasm.mod.ExportedFunction("Setup")
	}

	wasm.cbTeardown = wasm.mod.ExportedFunction("teardown")
	if wasm.cbTeardown == nil {
		wasm.cbTeardown = wasm.mod.ExportedFunction("Teardown")
	}

	wasm.cbUpdate = wasm.mod.ExportedFunction("update")
	if wasm.cbUpdate == nil {
		wasm.cbUpdate = wasm.mod.ExportedFunction("Update")
	}

	wasm.cbRender = wasm.mod.ExportedFunction("render")
	if wasm.cbRender == nil {
		wasm.cbRender = wasm.mod.ExportedFunction("Render")
	}

	return wasm, nil
}

func (w *WasmRuntime) Teardown() {
	w.CallTeardown()
	_ = w.mod.Close(w.ctx)
	_ = w.compiled.Close(w.ctx)
	_ = w.rt.Close(w.ctx)
}

func (w *WasmRuntime) Reload(transferMemory bool) error {
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

	// Hold on to old memory while we load the new module
	var (
		oldMemory []byte
		oldSize   = w.mod.Memory().Size()
	)

	if transferMemory {
		var ok bool

		oldMemory, ok = w.mod.Memory().Read(0, oldSize)
		if !ok {
			return errors.New("unable to transfer old memory to new module")
		}
	}

	err = w.mod.Close(w.ctx)
	if err != nil {
		LogWarn("wasm - unable to close original module: %s", err)
	}

	w.mod = newMod

	if transferMemory {
		// Grow new module to hold old memory
		_, ok := w.mod.Memory().Grow(oldSize / 65536)
		if !ok {
			return errors.New("unable to resize new module memory")
		}

		// Finally give new module old memory
		ok = w.mod.Memory().Write(0, oldMemory)
		if !ok {
			return errors.New("unable to transfer module memory")
		}
	}

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

func (w *WasmRuntime) ConvertAndExpose(exportName string, proc any, wrapper api.GoModuleFunc) {
	t := reflect.TypeOf(proc)
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("Expose expects a function, was given: %s", t.Kind()))
	}

	var (
		args = make([]WasmType, 0)
		rets = make([]WasmType, 0)
	)

	var kindToType func(t reflect.Type) []WasmType

	kindToType = func(t reflect.Type) []WasmType {
		switch t.Kind() {
		case reflect.Float32:
			return []WasmType{WasmF32}
		case reflect.Int32, reflect.Int:
			return []WasmType{WasmI32}
		case reflect.Uint32, reflect.Uint, reflect.Bool:
			return []WasmType{WasmU32}
		case reflect.String:
			return []WasmType{WasmU32, WasmU32}
		case reflect.Struct:
			types := make([]WasmType, 0)
			for i := 0; i < t.NumField(); i += 1 {
				field := t.Field(i)
				if field.IsExported() {
					types = append(types, kindToType(field.Type)...)
				}
			}

			return types
		default:
			panic(fmt.Sprintf("value for function is unsupported %s", t))
		}
	}

	in := t.NumIn()
	for i := 0; i < in; i += 1 {
		arg := t.In(i)
		args = append(args, kindToType(arg)...)
	}

	out := t.NumOut()
	for i := 0; i < out; i += 1 {
		ret := t.Out(i)
		rets = append(rets, kindToType(ret)...)
	}

	LogDebug("wasm - export %s", exportName)
	w.host.NewFunctionBuilder().WithGoModuleFunction(wrapper, args, rets).Export(exportName)
}

func (w *WasmRuntime) Expose(ns string, proc api.GoModuleFunc, args, rets []WasmType) {
	t := reflect.TypeOf(proc)
	if t.Kind() != reflect.Func {
		panic(fmt.Sprintf("Expose expects a function, was given: %s", t.Kind()))
	}

	fn := runtime.FuncForPC(reflect.ValueOf(proc).Pointer())
	name, _ := strings.CutPrefix(fn.Name(), "main.wasm")
	export := ns + name

	LogDebug("wasm - exporting %s", export)

	w.host.NewFunctionBuilder().WithGoModuleFunction(proc, args, rets).Export(export)
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

func boolToU32(b bool) (r uint32) {
	if b {
		r = 1
	} else {
		r = 0
	}

	return
}

func u32ToBool(u uint32) bool {
	return u == 1
}

// Code generated by 'go generate ./...'; DO NOT EDIT.
package engine

import (
	"context"
	"github.com/tetratelabs/wazero/api"
)

func (a *Config) Expose(wasm *WasmRuntime) {
	wasm.ConvertAndExpose("ConfigGetEngineFlags", a.GetEngineFlags, wasmGetEngineFlags)
	wasm.ConvertAndExpose("ConfigSetEngineFlags", a.SetEngineFlags, wasmSetEngineFlags)

}

// Wasm wrappers for Config

// Calls Config.GetEngineFlags
func wasmGetEngineFlags(ctx context.Context, m api.Module, stack []WasmValue) {
	r0 := brut.Config.GetEngineFlags()
	stack[0] = api.EncodeU32(uint32(r0))
}

// Calls Config.SetEngineFlags
func wasmSetEngineFlags(ctx context.Context, m api.Module, stack []WasmValue) {
	arg0 := api.DecodeU32(stack[0])
	brut.Config.SetEngineFlags(
		uint32(arg0),
	)
}

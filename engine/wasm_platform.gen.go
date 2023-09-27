// Code generated by 'go generate ./...'; DO NOT EDIT.
package engine

import (
	"context"
	"github.com/tetratelabs/wazero/api"
)

func (a *Platform) Expose(wasm *WasmRuntime) {
	wasm.ConvertAndExpose("PlatformExit", a.Exit, wasmExit)
	wasm.ConvertAndExpose("PlatformFps", a.Fps, wasmFps)
	wasm.ConvertAndExpose("PlatformLog", a.Log, wasmLog)
	wasm.ConvertAndExpose("PlatformSetScreenSize", a.SetScreenSize, wasmSetScreenSize)
	wasm.ConvertAndExpose("PlatformTps", a.Tps, wasmTps)

}

// Wasm wrappers for Platform

// Calls Platform.Exit
func wasmExit(ctx context.Context, m api.Module, stack []WasmValue) {
	brut.Platform.Exit()
}

// Calls Platform.Fps
func wasmFps(ctx context.Context, m api.Module, stack []WasmValue) {
	r0 := brut.Platform.Fps()
	stack[0] = api.EncodeF32(float32(r0))
}

// Calls Platform.Log
func wasmLog(ctx context.Context, m api.Module, stack []WasmValue) {
	arg0_0 := api.DecodeU32(stack[0])
	arg0_1 := api.DecodeU32(stack[1])
	brut.Platform.Log(
		readWasmString(m.Memory(), arg0_0, arg0_1),
	)
}

// Calls Platform.SetScreenSize
func wasmSetScreenSize(ctx context.Context, m api.Module, stack []WasmValue) {
	arg0 := api.DecodeI32(stack[0])
	arg1 := api.DecodeI32(stack[1])
	brut.Platform.SetScreenSize(
		int32(arg0),
		int32(arg1),
	)
}

// Calls Platform.Tps
func wasmTps(ctx context.Context, m api.Module, stack []WasmValue) {
	r0 := brut.Platform.Tps()
	stack[0] = api.EncodeF32(float32(r0))
}

package main

type config struct {
	SetFlags func(flags ConfigFlag) `wasm:"ConfigSetFlags"`
}

type ConfigFlag uint32

const (
	// ConfigHotReloading enables auto reloading of the core wasm module
	ConfigHotReloading ConfigFlag = 1 << iota

	// ConfigEngineLogging enables internal engine logging messages
	ConfigEngineLogging
)

func ConfigSetFlags(flags ConfigFlag) {
	B.flags = flags
}

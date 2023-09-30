# BrutEngine Odin

This repo includes [Odin](https://odin-lang.org) bindings for BrutEngine. These bindings are only expected to run under the `freestanding_wasm32` target.

## Usage

To generate the bindings, `cd` into the `generate` directory and run `odin run .`

Now the bindings can be imported like so:

```odin
// This should be built with -target:freestanding_wasm32

import brut "brutengine_odin"

@export setup :: proc "c" () {
   brut.PlatformLog("Within setup!")
}

@export teardown :: proc "c" () {
   brut.PlatformLog("Within teardown!")
}

@export update :: proc "c" () {
   if brut.InputPressed(.Escape) {
      brut.PlatformExit()
   }
}

@export render :: proc "c" () {
   brut.GraphicsClear({ 1, 1, 1, 1 }) // White
}
```

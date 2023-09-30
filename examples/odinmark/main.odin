package main

import brut "brutengine_odin"

foo: brut.Texture

@(export)
setup :: proc "c" () {
   foo = brut.Texture(10)
   brut.PlatformLog("Setup")
}

@(export)
teardown :: proc "c" () {
   brut.PlatformLog("Teardown")
}

@(export)
update :: proc "c" () {
   if brut.InputPressed(.Escape) {
      brut.PlatformExit()
   }
}

@(export)
render :: proc "c" () {
   brut.GraphicsClear({ 1, 1, 1, 1 })
}

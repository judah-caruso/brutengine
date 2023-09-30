#!/usr/bin/env sh

odin build . -target:freestanding_wasm32 -o:speed -out:../../game.wasm -show-timings

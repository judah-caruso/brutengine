package main

import (
	"bytes"
	"context"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tetratelabs/wazero/api"
	"image"
	"os"
)

type asset struct {
	LoadImage func(path string) Texture `wasm:"AssetLoadImage" wrapper:"wasmLoadImage"`
}

// Texture is a non-zero handle
type (
	Texture     uint32
	textureData struct {
		handle *ebiten.Image
		path   string
	}
)

const InvalidTexture Texture = 0

func AssetLoadImage(path string) Texture {
	if id, ok := B.GetTextureByName(path); ok {
		return id
	}

	LogInfo("asset - loading texture %q", path)

	data, err := os.ReadFile(path)
	if err != nil {
		LogError("asset - unable to load texture %q", path)
		return InvalidTexture
	}

	buf := bytes.NewBuffer(data)

	decoded, _, err := image.Decode(buf)
	if err != nil {
		LogError("asset - unable to decode texture %q! %s", path, err)
		return InvalidTexture
	}

	img := ebiten.NewImageFromImage(decoded)
	if img == nil {
		LogError("asset - unable to create image from texture %q", path)
		return InvalidTexture
	}

	id := Texture(len(B.loadedTextures) + 1)

	B.loadedTextures[id] = textureData{
		handle: img,
		path:   path,
	}

	LogInfo("asset - texture loaded!")
	return id
}

func wasmLoadImage(_ context.Context, m api.Module, offset, count uint32) uint32 {
	path := readWasmString(m.Memory(), offset, count)
	return uint32(AssetLoadImage(path))
}

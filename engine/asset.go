package engine

import (
	"bytes"
	"image"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

type (
	Asset struct {
		loadedTextures map[Texture]textureData
	}
	IAsset interface {
		LoadTexture(name string) Texture
	}

	// Texture is a non-zero texture id that can be used to get textureData
	Texture uint32

	// textureData is the internal representation of a texture
	textureData struct {
		name   string
		handle *ebiten.Image
	}
)

// InvalidTexture is used to signal when a texture was unable to be loaded or fetched
const InvalidTexture Texture = 0

func (a *Asset) getTextureByName(name string) (Texture, bool) {
	for id, data := range a.loadedTextures {
		if data.name == name {
			return id, true
		}
	}

	return InvalidTexture, false
}

func (a *Asset) GetTextureData(tex Texture) (*ebiten.Image, bool) {
	data, ok := a.loadedTextures[tex]
	if ok {
		return data.handle, true
	}

	return nil, false
}

func (a *Asset) Setup() error {
	a.loadedTextures = make(map[Texture]textureData)
	return nil
}

func (a *Asset) LoadTexture(name string) Texture {
	if id, ok := a.getTextureByName(name); ok {
		return id
	}

	LogDebug("asset - loading texture %q", name)

	data, err := os.ReadFile(name)
	if err != nil {
		LogError("asset - unable to load texture %q", name)
		return InvalidTexture
	}

	buf := bytes.NewBuffer(data)

	decoded, _, err := image.Decode(buf)
	if err != nil {
		LogError("asset - unable to decode texture %q! %s", name, err)
		return InvalidTexture
	}

	img := ebiten.NewImageFromImage(decoded)
	if img == nil {
		LogError("asset - unable to create image from texture %q", name)
		return InvalidTexture
	}

	id := Texture(len(a.loadedTextures) + 1)

	a.loadedTextures[id] = textureData{
		name:   name,
		handle: img,
	}

	LogDebug("asset - texture loaded!")
	return id
}

var _ IAsset = (*Asset)(nil)

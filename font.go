package main

import (
	"fmt"
	"image"
	"io/ioutil"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Font struct {
	face     font.Face
	textures map[rune]uint32
}

func NewFont(path string, size int) (*Font, error) {
	fontBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read font file: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse font: %v", err)
	}

	face := truetype.NewFace(f, &truetype.Options{
		Size:    float64(size),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	return &Font{
		face:     face,
		textures: make(map[rune]uint32),
	}, nil
}

func (f *Font) GetCharTexture(char rune) uint32 {
	if texture, ok := f.textures[char]; ok {
		return texture
	}

	// Create a new texture for this character
	bounds, advance, ok := f.face.GlyphBounds(char)
	if !ok {
		return 0
	}

	width := int(advance.Round())
	height := bounds.Max.Y.Round() - bounds.Min.Y.Round()

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	d := &font.Drawer{
		Dst:  img,
		Src:  image.White,
		Face: f.face,
		Dot:  fixed.Point26_6{X: -bounds.Min.X, Y: -bounds.Min.Y},
	}
	d.DrawString(string(char))

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(width), int32(height), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(img.Pix))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.BindTexture(gl.TEXTURE_2D, 0)

	f.textures[char] = texture
	return texture
}

func (f *Font) Destroy() {
	for _, texture := range f.textures {
		gl.DeleteTextures(1, &texture)
	}
}

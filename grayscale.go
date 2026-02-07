package phash

import (
	"image"

	"golang.org/x/image/draw"
)

// Grayscale converts any image.Image to *image.Gray.
// Uses standard luminance conversion (sRGB).
func Grayscale(src image.Image) *image.Gray {
	if src == nil {
		return nil
	}

	b := src.Bounds()
	dst := image.NewGray(image.Rect(0, 0, b.Dx(), b.Dy()))

	// draw.Draw handles color model conversion for us.
	draw.Draw(dst, dst.Bounds(), src, b.Min, draw.Src)

	return dst
}

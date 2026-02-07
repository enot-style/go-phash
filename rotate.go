package phash

import (
	"image"
	"image/color"
)

// ---------- Orientation transforms ----------
//
// Implementation notes:
// - All outputs are *image.RGBA with Bounds() starting at (0,0).
// - Fast path for *image.RGBA/*image.NRGBA using Pix copy instead of Set/At.
//
// If you want even more speed, you can add additional fast paths for *image.YCbCr,
// but for most ingestion pipelines the below is plenty and keeps code simple.

func flipHorizontal(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	if src, ok := img.(*image.RGBA); ok {
		flipHorizontalRGBA(dst, src, b)
		return dst
	}
	if src, ok := img.(*image.NRGBA); ok {
		flipHorizontalNRGBA(dst, src, b)
		return dst
	}

	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		for x := 0; x < w; x++ {
			sx := b.Min.X + (w - 1 - x)
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}

func flipVertical(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	if src, ok := img.(*image.RGBA); ok {
		flipVerticalRGBA(dst, src, b)
		return dst
	}
	if src, ok := img.(*image.NRGBA); ok {
		flipVerticalNRGBA(dst, src, b)
		return dst
	}

	for y := 0; y < h; y++ {
		sy := b.Min.Y + (h - 1 - y)
		for x := 0; x < w; x++ {
			sx := b.Min.X + x
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}

func rotate180(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	if src, ok := img.(*image.RGBA); ok {
		rotate180RGBA(dst, src, b)
		return dst
	}
	if src, ok := img.(*image.NRGBA); ok {
		rotate180NRGBA(dst, src, b)
		return dst
	}

	for y := 0; y < h; y++ {
		sy := b.Min.Y + (h - 1 - y)
		for x := 0; x < w; x++ {
			sx := b.Min.X + (w - 1 - x)
			dst.Set(x, y, img.At(sx, sy))
		}
	}
	return dst
}

// rotate90 rotates the image 90 degrees clockwise.
func rotate90(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	if src, ok := img.(*image.RGBA); ok {
		rotate90RGBA(dst, src, b)
		return dst
	}
	if src, ok := img.(*image.NRGBA); ok {
		rotate90NRGBA(dst, src, b)
		return dst
	}

	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		for x := 0; x < w; x++ {
			sx := b.Min.X + x
			// dst(x', y') where x' in [0,h), y' in [0,w)
			// Your original mapping: dst.Set(h-1-y, x, src(x,y))
			dst.Set(h-1-y, x, img.At(sx, sy))
		}
	}
	return dst
}

// rotate270 rotates the image 270 degrees clockwise (90 degrees counterclockwise).
func rotate270(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	if src, ok := img.(*image.RGBA); ok {
		rotate270RGBA(dst, src, b)
		return dst
	}
	if src, ok := img.(*image.NRGBA); ok {
		rotate270NRGBA(dst, src, b)
		return dst
	}

	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		for x := 0; x < w; x++ {
			sx := b.Min.X + x
			// Your original mapping: dst.Set(y, w-1-x, src(x,y))
			dst.Set(y, w-1-x, img.At(sx, sy))
		}
	}
	return dst
}

// transpose corresponds to EXIF orientation 5.
// It mirrors across the main diagonal: dst(x,y) = src(y,x)
func transpose(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		for x := 0; x < w; x++ {
			sx := b.Min.X + x
			// dst(y,x) = src(x,y)
			dst.Set(y, x, img.At(sx, sy))
		}
	}
	return dst
}

// transverse corresponds to EXIF orientation 7.
// It mirrors across the anti-diagonal: dst(x,y) = src(w-1-y, h-1-x)
func transverse(img image.Image) *image.RGBA {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, h, w))

	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		for x := 0; x < w; x++ {
			sx := b.Min.X + x
			// dst(h-1-y, w-1-x) = src(x,y)  (equivalent anti-diagonal reflection)
			dst.Set(h-1-y, w-1-x, img.At(sx, sy))
		}
	}
	return dst
}

// ---------- Fast paths for RGBA/NRGBA ----------

func flipHorizontalRGBA(dst, src *image.RGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			// src x mirrored
			mx := w - 1 - x
			so := sOff + mx*4
			do := dst.PixOffset(x, y)
			copy(dst.Pix[do:do+4], src.Pix[so:so+4])
		}
	}
}

func flipHorizontalNRGBA(dst *image.RGBA, src *image.NRGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			mx := w - 1 - x
			so := sOff + mx*4
			do := dst.PixOffset(x, y)
			// Convert NRGBA -> RGBA (un-premultiply).
			c := color.NRGBA{R: src.Pix[so], G: src.Pix[so+1], B: src.Pix[so+2], A: src.Pix[so+3]}
			r := color.RGBAModel.Convert(c).(color.RGBA)
			dst.Pix[do], dst.Pix[do+1], dst.Pix[do+2], dst.Pix[do+3] = r.R, r.G, r.B, r.A
		}
	}
}

func flipVerticalRGBA(dst, src *image.RGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + (h - 1 - y)
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			so := sOff + x*4
			do := dst.PixOffset(x, y)
			copy(dst.Pix[do:do+4], src.Pix[so:so+4])
		}
	}
}

func flipVerticalNRGBA(dst *image.RGBA, src *image.NRGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + (h - 1 - y)
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			so := sOff + x*4
			do := dst.PixOffset(x, y)
			c := color.NRGBA{R: src.Pix[so], G: src.Pix[so+1], B: src.Pix[so+2], A: src.Pix[so+3]}
			r := color.RGBAModel.Convert(c).(color.RGBA)
			dst.Pix[do], dst.Pix[do+1], dst.Pix[do+2], dst.Pix[do+3] = r.R, r.G, r.B, r.A
		}
	}
}

func rotate180RGBA(dst, src *image.RGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + (h - 1 - y)
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			mx := w - 1 - x
			so := sOff + mx*4
			do := dst.PixOffset(x, y)
			copy(dst.Pix[do:do+4], src.Pix[so:so+4])
		}
	}
}

func rotate180NRGBA(dst *image.RGBA, src *image.NRGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + (h - 1 - y)
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			mx := w - 1 - x
			so := sOff + mx*4
			do := dst.PixOffset(x, y)
			c := color.NRGBA{R: src.Pix[so], G: src.Pix[so+1], B: src.Pix[so+2], A: src.Pix[so+3]}
			r := color.RGBAModel.Convert(c).(color.RGBA)
			dst.Pix[do], dst.Pix[do+1], dst.Pix[do+2], dst.Pix[do+3] = r.R, r.G, r.B, r.A
		}
	}
}

func rotate90RGBA(dst, src *image.RGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	// dst is (h, w)
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			so := sOff + x*4
			dx, dy := h-1-y, x
			do := dst.PixOffset(dx, dy)
			copy(dst.Pix[do:do+4], src.Pix[so:so+4])
		}
	}
}

func rotate90NRGBA(dst *image.RGBA, src *image.NRGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			so := sOff + x*4
			dx, dy := h-1-y, x
			do := dst.PixOffset(dx, dy)
			c := color.NRGBA{R: src.Pix[so], G: src.Pix[so+1], B: src.Pix[so+2], A: src.Pix[so+3]}
			r := color.RGBAModel.Convert(c).(color.RGBA)
			dst.Pix[do], dst.Pix[do+1], dst.Pix[do+2], dst.Pix[do+3] = r.R, r.G, r.B, r.A
		}
	}
}

func rotate270RGBA(dst, src *image.RGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			so := sOff + x*4
			dx, dy := y, w-1-x
			do := dst.PixOffset(dx, dy)
			copy(dst.Pix[do:do+4], src.Pix[so:so+4])
		}
	}
}

func rotate270NRGBA(dst *image.RGBA, src *image.NRGBA, b image.Rectangle) {
	w, h := b.Dx(), b.Dy()
	for y := 0; y < h; y++ {
		sy := b.Min.Y + y
		sOff := src.PixOffset(b.Min.X, sy)
		for x := 0; x < w; x++ {
			so := sOff + x*4
			dx, dy := y, w-1-x
			do := dst.PixOffset(dx, dy)
			c := color.NRGBA{R: src.Pix[so], G: src.Pix[so+1], B: src.Pix[so+2], A: src.Pix[so+3]}
			r := color.RGBAModel.Convert(c).(color.RGBA)
			dst.Pix[do], dst.Pix[do+1], dst.Pix[do+2], dst.Pix[do+3] = r.R, r.G, r.B, r.A
		}
	}
}

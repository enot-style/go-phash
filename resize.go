package phash

import (
	"image"
	"math"

	"golang.org/x/image/draw"
)

// ResizeMultiStep resizes src to (dstW, dstH) using a quality-preserving strategy.
//
// - If dstW==0 or dstH==0, aspect ratio is preserved.
// - Downscale: progressive halving using CatmullRom, then final CatmullRom to exact size.
// - Upscale: single ApproxBiLinear pass (smoother, fewer halos).
//
// Official repos only: stdlib + golang.org/x/image/draw.
func Resize(src image.Image, dstW, dstH uint32) image.Image {
	if src == nil {
		return nil
	}

	sb := src.Bounds()
	sw, sh := sb.Dx(), sb.Dy()
	swu := uint32(sw)
	shu := uint32(sh)

	if (dstW == 0 && dstH == 0) || (dstW == swu && dstH == shu) {
		return src
	}

	// Preserve aspect ratio if one side is 0.
	if dstW == 0 {
		dstW = uint32(float64(dstH) * float64(sw) / float64(sh))
	}
	if dstH == 0 {
		dstH = uint32(float64(dstW) * float64(sh) / float64(sw))
	}

	if dstW == 0 || dstH == 0 {
		return src
	}

	// Upscale: smoother filter to avoid ringing/halos.
	if dstW >= swu && dstH >= shu {
		dst := image.NewRGBA(image.Rect(0, 0, int(dstW), int(dstH)))
		draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, sb, draw.Over, nil)
		return dst
	}

	// Downscale: progressive halving for quality.
	cur := src
	cw, ch := sw, sh

	for cw/2 >= int(dstW) && ch/2 >= int(dstH) {
		nw, nh := cw/2, ch/2
		tmp := image.NewRGBA(image.Rect(0, 0, nw, nh))
		draw.CatmullRom.Scale(tmp, tmp.Bounds(), cur, cur.Bounds(), draw.Over, nil)
		cur = tmp
		cw, ch = nw, nh
	}

	dst := image.NewRGBA(image.Rect(0, 0, int(dstW), int(dstH)))
	draw.CatmullRom.Scale(dst, dst.Bounds(), cur, cur.Bounds(), draw.Over, nil)
	return dst
}

// DownscaleByLargestSide scales the image down so the largest side is at most maxSide,
// preserving aspect ratio. If no downscale is needed, it returns src.
// If src is nil or maxSide is 0, it returns src.
func DownscaleByLargestSide(src image.Image, maxSide uint32) image.Image {
	if src == nil || maxSide == 0 {
		return src
	}

	b := src.Bounds()
	origW := uint32(b.Dx())
	origH := uint32(b.Dy())
	if origW == 0 || origH == 0 {
		return src
	}

	largest := origW
	if origH > largest {
		largest = origH
	}
	if largest <= maxSide {
		return src
	}

	scale := float64(maxSide) / float64(largest)
	newW := uint32(math.Round(float64(origW) * scale))
	newH := uint32(math.Round(float64(origH) * scale))
	if newW == 0 {
		newW = 1
	}
	if newH == 0 {
		newH = 1
	}
	return Resize(src, newW, newH)
}

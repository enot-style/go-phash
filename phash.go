package phash

import (
	"image"
	"math"
	"math/bits"
	"sort"
)

// PHash computes a classic 64-bit perceptual hash (pHash).
//
// Pipeline (classic 64-bit):
//  1. Resize to 32x32
//  2. Grayscale
//  3. 2D DCT (N=32), keep top-left 8x8 coefficients
//  4. Median of 63 coefficients excluding DC
//  5. Build 64-bit hash: bit=1 if coeff>median, with DC bit forced to 0
func PHash(image image.Image) uint64 {
	if image == nil {
		return 0
	}
	gray := Grayscale(image)
	resized := Resize(gray, 32, 32)
	pix := gray32x32(resized)                  // [32][32]float64 (0..255)
	coeff := dctTopLeft8x8(pix)                // [8][8]float64
	med := medianImageHash(coeff)              // float64
	return hashFromCoeffsImageHash(coeff, med) // uint64
}

// HammingDistance returns the number of differing bits between two 64-bit hashes.
func HammingDistance(a, b uint64) int { return bits.OnesCount64(a ^ b) }

func gray32x32(img image.Image) [32][32]float64 {
	var out [32][32]float64
	b := img.Bounds()
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			r, _, _, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
			out[y][x] = float64(r >> 8)
		}
	}
	return out
}

// Precomputed cosine table for N=32:
// cos32[k][n] = cos((2*n+1)*k*pi/(2*N)), where k in [0..7], n in [0..31]
var cos32 = func() [8][32]float64 {
	const N = 32.0
	var t [8][32]float64
	for k := 0; k < 8; k++ {
		for n := 0; n < 32; n++ {
			t[k][n] = math.Cos((2*float64(n) + 1.0) * float64(k) * math.Pi / (2.0 * N))
		}
	}
	return t
}()

// dctTopLeft8x8 computes the top-left 8x8 DCT coefficients from a 32x32 block of pixel values.
func dctTopLeft8x8(pix [32][32]float64) [8][8]float64 {
	const N = 32.0

	var c [8][8]float64
	for u := range 8 {
		au := math.Sqrt(2.0 / N)
		if u == 0 {
			au = math.Sqrt(1.0 / N)
		}
		for v := 0; v < 8; v++ {
			av := math.Sqrt(2.0 / N)
			if v == 0 {
				av = math.Sqrt(1.0 / N)
			}
			var sum float64
			for y := 0; y < 32; y++ {
				cvy := cos32[v][y]
				for x := 0; x < 32; x++ {
					sum += pix[y][x] * cos32[u][x] * cvy
				}
			}
			// keep your transpose so [yfreq][xfreq] to match ImageHash's [row][col]
			c[v][u] = au * av * sum
		}
	}
	return c
}

// medianImageHash computes the median of the 63 DCT coefficients excluding the DC component at c[0][0].
func medianImageHash(c [8][8]float64) float64 {
	// ImageHash uses median of c[1:, 1:]
	v := make([]float64, 0, 49)
	for y := 1; y < 8; y++ {
		for x := 1; x < 8; x++ {
			v = append(v, c[y][x])
		}
	}
	sort.Float64s(v)
	return v[len(v)/2]
}

// hashFromCoeffsImageHash builds the 64-bit hash from the DCT coefficients and median.
// Bit=1 if coeff>median, with DC bit forced to 0.
func hashFromCoeffsImageHash(c [8][8]float64, med float64) uint64 {
	// Flatten row-major: y then x, MSB-first (matches ImageHash hex output)
	var h uint64
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			h <<= 1
			if c[y][x] > med {
				h |= 1
			}
		}
	}
	return h
}

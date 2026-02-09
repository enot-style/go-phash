# go-phash

A small, dependency-light Go library and CLI for computing 64-bit perceptual hashes (pHash) of images. Useful for near-duplicate detection, similarity search, and basic image de-duplication workflows.

**Highlights**
- Classic 64-bit pHash pipeline (32x32 resize → grayscale → DCT → median threshold → 64-bit hash).
- CLI that hashes a file/URL or compares two images with Hamming distance.
- Robust decoding helpers with JPEG EXIF orientation handling.
- Built-in WebP support (decode and lossless encode).
- Simple image utilities (grayscale, resize, downscale).

**Install**
```bash
go get github.com/enot-style/go-phash
```

**CLI Usage**
Hash a single image (file path or URL):
```bash
go run ./cmd/phash test_data/sweater-thumb.jpg
```
Output is a 16-hex-digit hash:
```
fa85955a872769cb
```

Compare two images and get Hamming distance (distance only):
```bash
go run ./cmd/phash image-a.jpg image-b.jpg
```
Output format (single line):
```
<distance>
```

Build the CLI:
```bash
go build -o phash ./cmd/phash
```

**Library Usage**
```go
package main

import (
	"fmt"
	"os"

	"github.com/enot-style/go-phash"
)

func main() {
	f, err := os.Open("test_data/sweater-thumb.jpg")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := phash.DecodeAny(f)
	if err != nil {
		panic(err)
	}

	h := phash.PHash(img)
	fmt.Printf("%016x\n", h)
}
```

**API Overview**
Core hashing:
- `PHash(image.Image) uint64` computes the 64-bit perceptual hash.
- `HammingDistance(a, b uint64) int` compares two hashes.

Decoding helpers:
- `DecodeAny(io.Reader) (image.Image, string, error)` reads all bytes, decodes, and applies JPEG EXIF orientation.
- `DownloadAndDecodeAny(context.Context, string) (image.Image, string, error)` fetches over HTTP and decodes.
- `DownloadAndDecodeAnyWithLimit(context.Context, string, int64) (image.Image, string, error)` with size cap.

Encoding helper:
- `EncodeWebPLossless(io.Writer, image.Image) error` encodes lossless WebP.

Image utilities:
- `Grayscale(image.Image) *image.Gray`
- `Resize(image.Image, uint32, uint32) image.Image`
- `DownscaleByLargestSide(image.Image, uint32) image.Image`

**Supported Image Formats**
Decode (registered by default):
- JPEG, PNG, GIF, WebP (via `golang.org/x/image/webp`).

Encode:
- WebP (lossless) via `EncodeWebPLossless`.

**EXIF Orientation**
When decoding JPEGs, EXIF orientation is applied automatically, so hashes are stable across rotated inputs.

**Testing**
```bash
go test ./...
```

**Notes**
- `PHash(nil)` returns `0`.
- Hashes are 64-bit values typically rendered as 16 hex characters with `%016x`.
- The CLI accepts `http://` and `https://` URLs as inputs.

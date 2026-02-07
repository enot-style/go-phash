package phash

import (
	"image"
	"os"
	"path/filepath"
	"testing"
)

func TestPHashSweaterVariantsMatchExpected(t *testing.T) {
	const expected uint64 = 0xfa85955a872769cb
	paths := []string{
		filepath.Join("test_data", "sweater-thumb.jpg"),
		filepath.Join("test_data", "sweater-medium.jpg"),
		filepath.Join("test_data", "sweater-large.jpg"),
	}

	for _, path := range paths {
		path := path
		t.Run(path, func(t *testing.T) {
			img := decodeTestImage(t, path)
			hash := PHash(img)
			if hash != expected {
				t.Fatalf("unexpected pHash for %s: got %016x want %016x", path, hash, expected)
			}
		})
	}
}

func decodeTestImage(t *testing.T, path string) image.Image {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()

	decoded, _, err := DecodeAny(f)
	if err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
	return decoded
}

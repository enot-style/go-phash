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

func TestPHashColorOnlyDifferencesStayClose(t *testing.T) {
	testCases := []struct {
		name  string
		left  string
		right string
		max   int
	}{
		{
			name:  "tblue_vs_tgray",
			left:  filepath.Join("test_data", "tblue.jpeg"),
			right: filepath.Join("test_data", "tgray.jpeg"),
			max:   6,
		},
		{
			name:  "kblue_vs_kyellow",
			left:  filepath.Join("test_data", "kblue.webp"),
			right: filepath.Join("test_data", "kyellow.jpeg"),
			max:   6,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			left := decodeTestImage(t, tc.left)
			right := decodeTestImage(t, tc.right)
			distance := HammingDistance(PHash(left), PHash(right))

			if distance > tc.max {
				t.Fatalf(
					"distance too high for %s and %s: got %d want <= %d",
					tc.left,
					tc.right,
					distance,
					tc.max,
				)
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

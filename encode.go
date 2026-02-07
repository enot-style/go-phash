package phash

import (
	"fmt"
	"image"
	"io"

	"github.com/HugoSmits86/nativewebp"
)

// EncodeWebPLossless encodes img as a *lossless* WebP (VP8L) stream using the pure-Go encoder.
// It returns EncodeError for nil writer/image or encoder failures.
func EncodeWebPLossless(w io.Writer, img image.Image) error {
	if w == nil {
		return EncodeError{Op: EncodeOpWebP, Err: fmt.Errorf("nil writer")}
	}
	if img == nil {
		return EncodeError{Op: EncodeOpWebP, Err: fmt.Errorf("nil image")}
	}

	// Passing nil options uses library defaults.
	if err := nativewebp.Encode(w, img, nil); err != nil {
		return EncodeError{Op: EncodeOpWebP, Err: err}
	}
	return nil
}

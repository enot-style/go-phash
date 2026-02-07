package phash

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"net/http"

	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/webp"
)

// DecodeAny reads all bytes (so it works with non-seekable readers), decodes, and applies EXIF orientation.
// It returns the decoded image and the detected format string ("jpeg", "png", "gif", "webp", ...).
// Errors are returned as DecodeError with Op "read" or "decode".
func DecodeAny(r io.Reader) (image.Image, string, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpRead, Err: err}
	}
	return decodeBytes(b)
}

// DownloadAndDecodeAny fetches a remote image over HTTP, decodes it, and applies EXIF orientation.
// Errors are returned as DecodeError with Op "request", "http", "http status", or "decode".
func DownloadAndDecodeAny(ctx context.Context, url string) (image.Image, string, error) {
	var (
		req *http.Request
		err error
	)
	if ctx == nil {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	}
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpRequest, Err: err}
	}
	req.Header.Set("Accept", "image/*,*/*;q=0.8")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpHTTP, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, "", DecodeError{Op: DecodeOpHTTPStatus, Err: fmt.Errorf("%d (%s)", resp.StatusCode, resp.Status)}
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpRead, Err: err}
	}
	return decodeBytes(b)
}

// DownloadAndDecodeAnyWithLimit fetches a remote image over HTTP, decodes it with a byte cap, and applies EXIF orientation.
// Errors are returned as DecodeError with Op "request", "http", "http status", or "decode".
func DownloadAndDecodeAnyWithLimit(ctx context.Context, url string, maxBytes int64) (image.Image, string, error) {
	var (
		req *http.Request
		err error
	)
	if ctx == nil {
		req, err = http.NewRequest(http.MethodGet, url, nil)
	} else {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	}
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpRequest, Err: err}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpHTTP, Err: err}
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return nil, "", DecodeError{Op: DecodeOpHTTPStatus, Err: fmt.Errorf("%d (%s)", resp.StatusCode, resp.Status)}
	}
	limited := io.LimitReader(resp.Body, maxBytes)
	b, err := io.ReadAll(limited)
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpRead, Err: err}
	}
	return decodeBytes(b)
}

// decodeBytes decodes an image from bytes and normalizes it using EXIF orientation (JPEG only).
// Errors are returned as DecodeError with Op "decode".
func decodeBytes(b []byte) (image.Image, string, error) {
	img, format, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		return nil, "", DecodeError{Op: DecodeOpDecode, Err: err}
	}
	return applyEXIFOrientation(img, b), format, nil
}

// applyEXIFOrientation returns an image rotated/flipped per EXIF orientation if present.
// It never returns errors; missing or invalid EXIF keeps the original image.
//
// Orientation values (EXIF):
//
//	1: normal
//	2: mirror horizontal
//	3: rotate 180
//	4: mirror vertical
//	5: transpose (mirror across main diagonal)
//	6: rotate 90 CW
//	7: transverse (mirror across anti-diagonal)
//	8: rotate 270 CW
func applyEXIFOrientation(img image.Image, payload []byte) image.Image {
	orientation, ok := exifOrientationJPEG(payload)
	if !ok || orientation == 1 {
		return img
	}
	switch orientation {
	case 2:
		return flipHorizontal(img)
	case 3:
		return rotate180(img)
	case 4:
		return flipVertical(img)
	case 5:
		return transpose(img)
	case 6:
		return rotate90(img)
	case 7:
		return transverse(img)
	case 8:
		return rotate270(img)
	default:
		return img
	}
}

// exifOrientationJPEG attempts to read EXIF orientation from a JPEG payload.
// It returns the orientation value (1..8) and true on success.
func exifOrientationJPEG(data []byte) (int, bool) {
	// JPEG SOI: FF D8
	if len(data) < 4 || data[0] != 0xFF || data[1] != 0xD8 {
		return 0, false
	}

	// Scan JPEG markers until SOS (0xDA) or EOI (0xD9).
	for i := 2; i+4 <= len(data); {
		// Find marker prefix 0xFF (can be padded).
		if data[i] != 0xFF {
			i++
			continue
		}
		if i+1 >= len(data) {
			return 0, false
		}

		marker := data[i+1]
		i += 2

		// EOI or SOS: no more metadata segments after SOS.
		if marker == 0xD9 || marker == 0xDA {
			break
		}

		// Standalone markers (no length field).
		if marker == 0x01 || (marker >= 0xD0 && marker <= 0xD7) {
			continue
		}

		// Need segment length.
		if i+2 > len(data) {
			break
		}
		segLen := int(binary.BigEndian.Uint16(data[i : i+2])) // includes these 2 bytes
		if segLen < 2 {
			break
		}
		segEnd := i + segLen
		if segEnd > len(data) {
			break
		}

		// APP1 (Exif) marker is 0xE1.
		if marker == 0xE1 && segLen >= 8 {
			segment := data[i+2 : segEnd] // exclude length bytes
			// "Exif\0\0" header followed by TIFF data
			if len(segment) >= 6 && bytes.HasPrefix(segment, exifHeader) {
				if orientation, ok := parseExifOrientation(segment[6:]); ok {
					return orientation, true
				}
			}
		}

		i = segEnd
	}

	return 0, false
}

var exifHeader = []byte("Exif\x00\x00")

// parseExifOrientation parses TIFF payload and extracts the Orientation tag if present.
// It returns the orientation value (1..8) and true on success.
func parseExifOrientation(tiff []byte) (int, bool) {
	if len(tiff) < 8 {
		return 0, false
	}

	var order binary.ByteOrder
	switch {
	case tiff[0] == 'I' && tiff[1] == 'I':
		order = binary.LittleEndian
	case tiff[0] == 'M' && tiff[1] == 'M':
		order = binary.BigEndian
	default:
		return 0, false
	}

	// TIFF magic number 42.
	if order.Uint16(tiff[2:4]) != 0x002A {
		return 0, false
	}

	ifdOffset := int(order.Uint32(tiff[4:8]))
	if ifdOffset < 0 || ifdOffset+2 > len(tiff) {
		return 0, false
	}

	entryCount := int(order.Uint16(tiff[ifdOffset : ifdOffset+2]))
	// Robustness cap: IFD0 rarely has many entries; prevents pathological loops.
	if entryCount < 0 || entryCount > 256 {
		return 0, false
	}

	entriesBase := ifdOffset + 2
	for n := 0; n < entryCount; n++ {
		entryOffset := entriesBase + n*12
		if entryOffset+12 > len(tiff) {
			break
		}

		tag := order.Uint16(tiff[entryOffset : entryOffset+2])
		if tag != 0x0112 { // Orientation
			continue
		}

		typ := order.Uint16(tiff[entryOffset+2 : entryOffset+4])
		count := order.Uint32(tiff[entryOffset+4 : entryOffset+8])

		// Orientation must be SHORT (3), count == 1.
		if typ != 3 || count != 1 {
			return 0, false
		}

		// For SHORT count==1, the value fits in the 4-byte value field.
		value := order.Uint16(tiff[entryOffset+8 : entryOffset+10])
		if value >= 1 && value <= 8 {
			return int(value), true
		}
		return 0, false
	}

	return 0, false
}

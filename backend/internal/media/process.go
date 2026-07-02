// Package media handles image processing for uploaded files: decode, resize,
// and re-encode originals + thumbnails as JPEG.
package media

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"

	_ "golang.org/x/image/webp"

	"github.com/disintegration/imaging"
)

const (
	MaxOriginalPx = 1600 // long edge cap for the optimised original
	MaxThumbPx    = 400  // long edge cap for the thumbnail
	JPEGQuality   = 82   // JPEG encoding quality
)

// Processed holds the results of processing one upload.
type Processed struct {
	OriginalBytes []byte
	ThumbBytes    []byte
	Width         int
	Height        int
}

// Process decodes src, enforces the pixel caps, and returns JPEG-encoded
// optimised original + thumbnail along with the original image dimensions.
// All output is JPEG regardless of the input format.
func Process(src io.Reader) (*Processed, error) {
	data, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("media: read: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("media: decode: %w", err)
	}

	bounds := img.Bounds()
	origW, origH := bounds.Dx(), bounds.Dy()

	optimised := imaging.Fit(img, MaxOriginalPx, MaxOriginalPx, imaging.Lanczos)
	thumb := imaging.Fit(img, MaxThumbPx, MaxThumbPx, imaging.Lanczos)

	origBytes, err := encodeJPEG(optimised)
	if err != nil {
		return nil, err
	}
	thumbBytes, err := encodeJPEG(thumb)
	if err != nil {
		return nil, err
	}

	return &Processed{
		OriginalBytes: origBytes,
		ThumbBytes:    thumbBytes,
		Width:         origW,
		Height:        origH,
	}, nil
}

func encodeJPEG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := imaging.Encode(&buf, img, imaging.JPEG, imaging.JPEGQuality(JPEGQuality)); err != nil {
		return nil, fmt.Errorf("media: encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

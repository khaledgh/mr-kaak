package media

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"
)

func TestProcess(t *testing.T) {
	// Create a simple 10x10 test image
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 0; y < 10; y++ {
		for x := 0; x < 10; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	// Minimal 1x1 transparent WebP image encoded in base64
	webpBase64 := "UklGRhIAAABXRUJQVlA4IBYAAAAwAQCdASoBAAEADsD+JaQAA3AAAAAA"
	webpBytes, err := base64.StdEncoding.DecodeString(webpBase64)
	if err != nil {
		t.Fatalf("failed to decode base64 webp: %v", err)
	}

	tests := []struct {
		name    string
		encode  func() ([]byte, error)
		wantW   int
		wantH   int
		wantErr bool
	}{
		{
			name: "JPEG image",
			encode: func() ([]byte, error) {
				var buf bytes.Buffer
				err := jpeg.Encode(&buf, img, nil)
				return buf.Bytes(), err
			},
			wantW:   10,
			wantH:   10,
			wantErr: false,
		},
		{
			name: "PNG image",
			encode: func() ([]byte, error) {
				var buf bytes.Buffer
				err := png.Encode(&buf, img)
				return buf.Bytes(), err
			},
			wantW:   10,
			wantH:   10,
			wantErr: false,
		},
		{
			name: "WebP image",
			encode: func() ([]byte, error) {
				return webpBytes, nil
			},
			wantW:   1,
			wantH:   1,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := tt.encode()
			if err != nil {
				t.Fatalf("failed to encode: %v", err)
			}

			res, err := Process(bytes.NewReader(data))
			if (err != nil) != tt.wantErr {
				t.Errorf("Process() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if res.Width != tt.wantW || res.Height != tt.wantH {
					t.Errorf("expected dimensions %dx%d, got %dx%d", tt.wantW, tt.wantH, res.Width, res.Height)
				}
				if len(res.OriginalBytes) == 0 || len(res.ThumbBytes) == 0 {
					t.Errorf("expected non-empty bytes")
				}
			}
		})
	}
}

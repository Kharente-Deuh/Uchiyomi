// SPDX-License-Identifier: AGPL-3.0-or-later

package download_test

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters/download"
)

func createTestPNG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png.Encode: %v", err)
	}

	return buf.Bytes()
}

func createTestJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 256), G: uint8(y % 256), B: 100, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}

	return buf.Bytes()
}

func createTestSolidJPEG(t *testing.T, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100}); err != nil {
		t.Fatalf("jpeg.Encode: %v", err)
	}

	return buf.Bytes()
}

func createTestAPNG(t *testing.T) []byte {
	t.Helper()
	basePNG := createTestPNG(t, 10, 10)
	// Insert an acTL chunk before IDAT chunk
	idatIdx := bytes.Index(basePNG, []byte("IDAT"))
	if idatIdx == -1 {
		t.Fatalf("IDAT chunk not found in test PNG")
	}
	chunkStart := idatIdx - 4

	var actlChunk bytes.Buffer
	// Length: 8 bytes (sequence_number: 4 bytes, num_plays: 4 bytes)
	_ = binary.Write(&actlChunk, binary.BigEndian, uint32(8))
	actlChunk.WriteString("acTL")
	_ = binary.Write(&actlChunk, binary.BigEndian, uint32(2)) // num_frames
	_ = binary.Write(&actlChunk, binary.BigEndian, uint32(0)) // num_plays
	// CRC (fake 4 bytes for testing detector)
	_ = binary.Write(&actlChunk, binary.BigEndian, uint32(0))

	var out bytes.Buffer
	out.Write(basePNG[:chunkStart])
	out.Write(actlChunk.Bytes())
	out.Write(basePNG[chunkStart:])

	return out.Bytes()
}

func TestSniffFormat(t *testing.T) {
	pngData := createTestPNG(t, 2, 2)
	jpegData := createTestJPEG(t, 2, 2)
	webpData := append([]byte("RIFF1234WEBP"), []byte("data")...)
	gifData := []byte("GIF89a...")
	corruptData := []byte("random non-image bytes")

	if download.SniffFormat(pngData) != download.FormatPNG {
		t.Errorf("expected FormatPNG, got %v", download.SniffFormat(pngData))
	}
	if download.SniffFormat(jpegData) != download.FormatJPEG {
		t.Errorf("expected FormatJPEG, got %v", download.SniffFormat(jpegData))
	}
	if download.SniffFormat(webpData) != download.FormatWebP {
		t.Errorf("expected FormatWebP, got %v", download.SniffFormat(webpData))
	}
	if download.SniffFormat(gifData) != download.FormatGIF {
		t.Errorf("expected FormatGIF, got %v", download.SniffFormat(gifData))
	}
	if download.SniffFormat(corruptData) != download.FormatUnknown {
		t.Errorf("expected FormatUnknown, got %v", download.SniffFormat(corruptData))
	}
}

func TestDetectExtension(t *testing.T) {
	pngData := createTestPNG(t, 2, 2)
	jpegData := createTestJPEG(t, 2, 2)
	webpData := append([]byte("RIFF1234WEBP"), []byte("data")...)
	gifData := []byte("GIF89a...")
	corruptData := []byte("random bytes")

	tests := []struct {
		name     string
		data     []byte
		urlExt   string
		expected string
	}{
		{name: "PNG data", data: pngData, urlExt: ".png", expected: download.ExtPNG},
		{name: "JPEG data with .jpg urlExt", data: jpegData, urlExt: ".jpg", expected: download.ExtJPG},
		{name: "JPEG data with .jpeg urlExt", data: jpegData, urlExt: ".jpeg", expected: download.ExtJPEG},
		{name: "WebP data", data: webpData, urlExt: ".webp", expected: download.ExtWEBP},
		{name: "GIF data", data: gifData, urlExt: ".gif", expected: download.ExtGIF},
		{name: "Unknown format fallback to urlExt", data: corruptData, urlExt: ".avif", expected: ".avif"},
		{name: "Unknown format empty urlExt defaults to webp", data: corruptData, urlExt: "", expected: download.ExtWEBP},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ext := download.DetectExtension(tc.data, tc.urlExt)
			if ext != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, ext)
			}
		})
	}
}

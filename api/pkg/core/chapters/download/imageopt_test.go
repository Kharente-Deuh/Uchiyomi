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

func TestIsAPNG(t *testing.T) {
	regularPNG := createTestPNG(t, 2, 2)
	if download.IsAPNG(regularPNG) {
		t.Errorf("regular PNG should not be detected as APNG")
	}

	apng := createTestAPNG(t)
	if !download.IsAPNG(apng) {
		t.Errorf("APNG with acTL chunk before IDAT should be detected as APNG")
	}
}

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

func TestOptimizePage_PNGToWebP(t *testing.T) {
	pngData := createTestPNG(t, 200, 200)
	res := download.OptimizePage(pngData, ".png", nil)

	if res.Extension != ".webp" {
		t.Errorf("expected extension .webp, got %s", res.Extension)
	}
	if len(res.Data) >= len(pngData) {
		t.Errorf("expected webp size (%d) to be smaller than png size (%d)", len(res.Data), len(pngData))
	}
	if download.SniffFormat(res.Data) != download.FormatWebP {
		t.Errorf("expected output to be valid WebP format")
	}
}

func TestOptimizePage_JPEGToWebP_Smaller(t *testing.T) {
	solidJPEG := createTestSolidJPEG(t, 200, 200)
	res := download.OptimizePage(solidJPEG, ".jpg", nil)

	if res.Extension != ".webp" {
		t.Errorf("expected .webp when smaller, got %s", res.Extension)
	}
	if len(res.Data) >= len(solidJPEG) {
		t.Errorf("expected webp size (%d) to be smaller than jpeg (%d)", len(res.Data), len(solidJPEG))
	}
	if download.SniffFormat(res.Data) != download.FormatWebP {
		t.Errorf("expected output to be valid WebP format")
	}

	gradientJPEG := createTestJPEG(t, 200, 200)
	resGradient := download.OptimizePage(gradientJPEG, ".jpg", nil)
	if len(resGradient.Data) < len(gradientJPEG) {
		if resGradient.Extension != ".webp" {
			t.Errorf("expected .webp when smaller, got %s", resGradient.Extension)
		}
	} else {
		if resGradient.Extension != ".jpg" {
			t.Errorf("expected original extension when WebP is larger, got %s", resGradient.Extension)
		}
		if !bytes.Equal(resGradient.Data, gradientJPEG) {
			t.Errorf("expected original data preserved when WebP is larger")
		}
	}
}

func TestOptimizePage_APNG_Preserved(t *testing.T) {
	apngData := createTestAPNG(t)
	res := download.OptimizePage(apngData, ".png", nil)

	if res.Extension != ".png" {
		t.Errorf("expected APNG to retain .png extension, got %s", res.Extension)
	}
	if !bytes.Equal(res.Data, apngData) {
		t.Errorf("expected APNG bytes to be preserved untouched")
	}
}

func TestOptimizePage_SourceWebP_Preserved(t *testing.T) {
	webpData := append([]byte("RIFF1234WEBP"), []byte("somedata")...)
	res := download.OptimizePage(webpData, ".webp", nil)

	if res.Extension != ".webp" {
		t.Errorf("expected source WebP to retain .webp extension, got %s", res.Extension)
	}
	if !bytes.Equal(res.Data, webpData) {
		t.Errorf("expected source WebP bytes to be preserved untouched")
	}
}

func TestOptimizePage_MagicBytesMismatch(t *testing.T) {
	// Content is PNG, but URL says .jpg
	pngData := createTestPNG(t, 200, 200)
	res := download.OptimizePage(pngData, ".jpg", nil)

	if res.Extension != ".webp" {
		t.Errorf("expected converted PNG to have .webp extension, got %s", res.Extension)
	}
	if download.SniffFormat(res.Data) != download.FormatWebP {
		t.Errorf("expected output to be WebP")
	}
}

func TestOptimizePage_CorruptData_Preserved(t *testing.T) {
	corruptData := []byte("not an image")
	res := download.OptimizePage(corruptData, ".png", nil)

	if res.Extension != ".png" {
		t.Errorf("expected corrupt data to retain url extension, got %s", res.Extension)
	}
	if !bytes.Equal(res.Data, corruptData) {
		t.Errorf("expected corrupt bytes to be preserved untouched")
	}
}

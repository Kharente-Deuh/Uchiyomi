// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/jpeg"
	"image/png"
	"log/slog"
	"strings"

	"github.com/KarpelesLab/gowebp"
)

type ImageFormat int

const (
	FormatUnknown ImageFormat = iota
	FormatPNG
	FormatJPEG
	FormatWebP
	FormatGIF
)

const (
	pngHeader  = "\x89PNG\r\n\x1a\n"
	jpegHeader = "\xff\xd8\xff"
	gif87a     = "GIF87a"
	gif89a     = "GIF89a"
)

type OptimizedPage struct {
	Data      []byte
	Extension string
}

func OptimizePage(data []byte, urlExt string, logger *slog.Logger) OptimizedPage {
	format := SniffFormat(data)
	defaultExt := fallbackExtension(format, urlExt)

	fallback := OptimizedPage{
		Data:      data,
		Extension: defaultExt,
	}

	if format == FormatPNG && IsAPNG(data) {
		return fallback
	}

	if format != FormatPNG && format != FormatJPEG {
		return fallback
	}

	var img image.Image
	var err error

	switch format {
	case FormatPNG:
		img, err = png.Decode(bytes.NewReader(data))
	case FormatJPEG:
		img, err = jpeg.Decode(bytes.NewReader(data))
	default:
		return fallback
	}

	if err != nil {
		if logger != nil {
			logger.Warn("failed to decode image for webp conversion", "error", err)
		}

		return fallback
	}

	var webpBuf bytes.Buffer
	// nil options in KarpelesLab/gowebp encodes lossless VP8L by default
	if err := gowebp.Encode(&webpBuf, img, nil); err != nil {
		if logger != nil {
			logger.Warn("failed to encode lossless webp", "error", err)
		}

		return fallback
	}

	webpBytes := webpBuf.Bytes()
	if len(webpBytes) < len(data) {
		return OptimizedPage{
			Data:      webpBytes,
			Extension: ".webp",
		}
	}

	return fallback
}

func fallbackExtension(format ImageFormat, urlExt string) string {
	switch format {
	case FormatPNG:
		return ".png"
	case FormatJPEG:
		if strings.EqualFold(urlExt, ".jpeg") {
			return ".jpeg"
		}

		return ".jpg"
	case FormatWebP:
		return ".webp"
	case FormatGIF:
		return ".gif"
	default:
		cleanExt := strings.ToLower(strings.TrimSpace(urlExt))
		if cleanExt == "" || !strings.HasPrefix(cleanExt, ".") {
			return ".webp"
		}

		return cleanExt
	}
}

func SniffFormat(data []byte) ImageFormat {
	if len(data) >= 8 && string(data[:8]) == pngHeader {
		return FormatPNG
	}

	if len(data) >= 3 && string(data[:3]) == jpegHeader {
		return FormatJPEG
	}

	if len(data) >= 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return FormatWebP
	}

	if len(data) >= 6 && (string(data[:6]) == gif87a || string(data[:6]) == gif89a) {
		return FormatGIF
	}

	return FormatUnknown
}

func IsAPNG(data []byte) bool {
	if len(data) < 8 || string(data[:8]) != pngHeader {
		return false
	}

	offset := 8
	for offset+8 <= len(data) {
		length := binary.BigEndian.Uint32(data[offset : offset+4])
		chunkType := string(data[offset+4 : offset+8])

		if chunkType == "acTL" {
			return true
		}

		if chunkType == "IDAT" {
			return false
		}

		// 4 (length) + 4 (type) + length (data) + 4 (crc)
		chunkTotal := int64(length) + 12
		if int64(offset)+chunkTotal > int64(len(data)) {
			break
		}

		offset += int(chunkTotal)
	}

	return false
}

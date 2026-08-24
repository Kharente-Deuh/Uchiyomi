// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"bytes"
	"encoding/binary"
)

type ImageFormat int

const (
	FormatUnknown ImageFormat = iota
	FormatPNG
	FormatJPEG
	FormatWebP
	FormatGIF
)

var pngHeader = []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
var jpegHeader = []byte{0xFF, 0xD8, 0xFF}
var gif87a = []byte("GIF87a")
var gif89a = []byte("GIF89a")

func SniffFormat(data []byte) ImageFormat {
	if len(data) >= 8 && bytes.Equal(data[:8], pngHeader) {
		return FormatPNG
	}

	if len(data) >= 3 && bytes.Equal(data[:3], jpegHeader) {
		return FormatJPEG
	}

	if len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) && bytes.Equal(data[8:12], []byte("WEBP")) {
		return FormatWebP
	}

	if len(data) >= 6 && (bytes.Equal(data[:6], gif87a) || bytes.Equal(data[:6], gif89a)) {
		return FormatGIF
	}

	return FormatUnknown
}

func IsAPNG(data []byte) bool {
	if len(data) < 8 || !bytes.Equal(data[:8], pngHeader) {
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

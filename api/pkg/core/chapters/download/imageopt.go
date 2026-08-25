// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"strings"
)

const (
	ExtWEBP = ".webp"
	ExtGIF  = ".gif"
	ExtJPG  = ".jpg"
	ExtJPEG = ".jpeg"
	ExtPNG  = ".png"
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

func DetectExtension(data []byte, urlExt string) string {
	format := SniffFormat(data)

	switch format {
	case FormatPNG:
		return ExtPNG
	case FormatJPEG:
		if strings.EqualFold(urlExt, ".jpeg") {
			return ExtJPEG
		}

		return ExtJPG
	case FormatWebP:
		return ExtWEBP
	case FormatGIF:
		return ExtGIF
	default:
		cleanExt := strings.ToLower(strings.TrimSpace(urlExt))
		if cleanExt == "" || !strings.HasPrefix(cleanExt, ".") {
			return ExtWEBP
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

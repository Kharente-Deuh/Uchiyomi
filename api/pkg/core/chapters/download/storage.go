// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
)

func comicDir(baseDir string, comicID uuid.UUID) string {
	return filepath.Join(baseDir, comicID.String())
}

func chapterDir(baseDir string, comicID uuid.UUID, chapterNumber float64) string {
	return filepath.Join(comicDir(baseDir, comicID), formatChapterNumber(chapterNumber))
}

func formatChapterNumber(number float64) string {
	return strconv.FormatFloat(number, 'f', -1, 64)
}

func pageFilename(pageIndex int, ext string) string {
	return fmt.Sprintf("%03d%s", pageIndex, ext)
}

func pageExtension(imageURL string) string {
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return ".webp"
	}

	ext := strings.ToLower(path.Ext(parsed.Path))
	if ext == "" {
		return ".webp"
	}

	return ext
}

func listDownloadedPageIndices(dir string) (map[int]struct{}, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return map[int]struct{}{}, nil
		}

		return nil, fmt.Errorf("os.ReadDir %s: %w", dir, err)
	}

	indices := make(map[int]struct{}, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		index, ok := parsePageIndex(entry.Name())
		if !ok {
			continue
		}

		indices[index] = struct{}{}
	}

	return indices, nil
}

func parsePageIndex(name string) (int, bool) {
	if len(name) < 4 {
		return 0, false
	}

	indexPart := name[:3]
	extPart := name[3:]
	if extPart == "" || !strings.HasPrefix(extPart, ".") {
		return 0, false
	}

	index, err := strconv.Atoi(indexPart)
	if err != nil || index <= 0 {
		return 0, false
	}

	return index, true
}

func deleteChapterDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("os.RemoveAll %s: %w", dir, err)
	}

	return nil
}

func deleteComicDir(dir string) error {
	err := os.RemoveAll(dir)
	if err != nil {
		return fmt.Errorf("os.RemoveAll %s: %w", dir, err)
	}

	return nil
}

type DiskPages struct {
	Dir string
}

func (s DiskPages) OpenPage(comicID uuid.UUID, chapterNumber float64, index int) (string, string, error) {
	pattern := filepath.Join(chapterDir(s.Dir, comicID, chapterNumber), pageFilename(index, ".*"))
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", "", fmt.Errorf("filepath.Glob %s: %w", pattern, err)
	}

	var files []string
	for _, match := range matches {
		st, err := os.Stat(match)
		if err != nil {
			return "", "", fmt.Errorf("os.Stat %s: %w", match, err)
		}

		if st.Mode().IsRegular() {
			files = append(files, match)
		}
	}

	if len(files) != 1 {
		return "", "", domain.ErrNotFound
	}

	contentType := mime.TypeByExtension(filepath.Ext(files[0]))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return files[0], contentType, nil
}

func downloadPage(
	ctx context.Context,
	client *http.Client,
	imageURL string,
	dir string,
	pageIndex int,
	_ *slog.Logger,
) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return fmt.Errorf("http.NewRequestWithContext: %w", err)
	}

	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("client.Do: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", res.StatusCode)
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("io.ReadAll: %w", err)
	}

	urlExt := pageExtension(imageURL)
	ext := DetectExtension(bodyBytes, urlExt)

	destPath := filepath.Join(dir, pageFilename(pageIndex, ext))

	if err = os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("os.MkdirAll %s: %w", dir, err)
	}

	tmp, err := os.CreateTemp(dir, ".dl-*")
	if err != nil {
		return fmt.Errorf("os.CreateTemp %s: %w", dir, err)
	}

	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if _, err = tmp.Write(bodyBytes); err != nil {
		return fmt.Errorf("tmp.Write %s: %w", tmpName, err)
	}

	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("tmp.Sync %s: %w", tmpName, err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("tmp.Close %s: %w", tmpName, err)
	}

	if err = os.Rename(tmpName, destPath); err != nil {
		return fmt.Errorf("os.Rename %s: %w", tmpName, err)
	}

	return nil
}

func progressPercent(downloadedPages, pagesNb int) int {
	if pagesNb <= 0 {
		return 0
	}

	progress := downloadedPages * 100 / pagesNb
	if progress > 100 {
		return 100
	}

	return progress
}

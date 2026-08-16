// SPDX-License-Identifier: AGPL-3.0-or-later

package covers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func (s *Service) ObtainLocal(ctx context.Context, comicID uuid.UUID, source, slug string) error {
	resolver, ok := s.deps.Resolvers[source]
	if !ok {
		return ErrUnknownSource
	}

	externalURL, err := resolver.ResolveExternalURL(ctx, slug)
	if err != nil {
		if errors.Is(err, ErrSeriesNotFound) {
			return ErrSeriesNotFound
		}

		return fmt.Errorf("resolver.ResolveExternalURL: %w", err)
	}

	ext := ExtensionFromURL(externalURL)
	if ext == "" {
		ext, err = probeExtension(ctx, s.deps.HTTPClient, externalURL)
		if err != nil {
			if errors.Is(err, ErrSeriesNotFound) {
				return ErrSeriesNotFound
			}

			return fmt.Errorf("%w: %w", ErrDownloadFailed, err)
		}
	}

	dest := filepath.Join(s.cfg.DownloadsDir, comicID.String(), "cover"+ext)

	moved, err := s.deps.Cache.Take(ctx, cacheKey(source, slug, ext), dest)
	if err != nil {
		return fmt.Errorf("cache.Take: %w", err)
	}

	if moved {
		return nil
	}

	rc, err := resolver.Fetch(ctx, externalURL)
	if err != nil {
		if errors.Is(err, ErrSeriesNotFound) {
			return ErrSeriesNotFound
		}

		return fmt.Errorf("%w: %w", ErrDownloadFailed, err)
	}

	defer rc.Close()

	if err = writeFileAtomic(dest, rc); err != nil {
		return err
	}

	return nil
}

func (s *Service) ServeLocal(_ context.Context, comicID uuid.UUID) (string, string, error) {
	dir := filepath.Join(s.cfg.DownloadsDir, comicID.String())
	matches, err := filepath.Glob(filepath.Join(dir, "cover.*"))
	if err != nil {
		return "", "", fmt.Errorf("filepath.Glob: %w", err)
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
		return "", "", ErrLocalCoverMissing
	}

	return files[0], MIMEForExtension(filepath.Ext(files[0])), nil
}

func (s *Service) RemoveLocal(comicID uuid.UUID) error {
	dir := filepath.Join(s.cfg.DownloadsDir, comicID.String())
	if err := os.RemoveAll(dir); err != nil {
		return fmt.Errorf("os.RemoveAll %s: %w", dir, err)
	}

	return nil
}

func writeFileAtomic(dest string, r io.Reader) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fmt.Errorf("os.MkdirAll %s: %w", dest, err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(dest), ".cover-*")
	if err != nil {
		return fmt.Errorf("os.CreateTemp %s: %w", dest, err)
	}

	tmpName := tmp.Name()
	defer func() {
		tmp.Close()
		os.Remove(tmpName)
	}()

	if _, err = io.Copy(tmp, r); err != nil {
		return fmt.Errorf("io.Copy %s: %w", tmpName, err)
	}

	if err = tmp.Sync(); err != nil {
		return fmt.Errorf("tmp.Sync %s: %w", tmpName, err)
	}

	if err = tmp.Close(); err != nil {
		return fmt.Errorf("tmp.Close %s: %w", tmpName, err)
	}

	if err = os.Rename(tmpName, dest); err != nil {
		return fmt.Errorf("os.Rename %s: %w", tmpName, err)
	}

	return nil
}

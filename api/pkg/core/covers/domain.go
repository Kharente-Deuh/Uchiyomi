// SPDX-License-Identifier: AGPL-3.0-or-later

package covers

import (
	"context"
	"errors"
	"io"

	"github.com/google/uuid"
)

const SourceAsuraScans = "asurascans"

var (
	ErrUnknownSource     = errors.New("unknown source")
	ErrSeriesNotFound    = errors.New("series not found")
	ErrDownloadFailed    = errors.New("cover download failed")
	ErrLocalCoverMissing = errors.New("local cover missing")
)

type CoverResolver interface {
	ResolveExternalURL(context.Context, string) (string, error)
	Fetch(context.Context, string) (io.ReadCloser, error)
}

type LocalComicFinder interface {
	FindBySourceSlug(ctx context.Context, source, slug string) (uuid.UUID, error)
}

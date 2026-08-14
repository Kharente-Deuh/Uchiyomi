// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"

	"github.com/google/uuid"
)

type ProgressStatus string

const (
	ProgressStatusDownloading ProgressStatus = "downloading"
	ProgressStatusCompleted   ProgressStatus = "completed"
	ProgressStatusError       ProgressStatus = "error"
)

type ProgressEvent struct {
	Status   ProgressStatus
	Download int
}

type ProgressPublisher interface {
	Publish(context.Context, uuid.UUID, ProgressEvent)
}

type noopPublisher struct{}

func (noopPublisher) Publish(context.Context, uuid.UUID, ProgressEvent) {}

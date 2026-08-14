// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"
	"slices"
	"sync"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

type queueItem struct {
	ChapterID uuid.UUID
	Number    float64
}

type queue struct {
	pending map[uuid.UUID]struct{}
	items   map[sources.SourceName][]queueItem
	notify  map[sources.SourceName]chan struct{}
	mu      sync.Mutex
}

func newQueue() *queue {
	return &queue{
		pending: make(map[uuid.UUID]struct{}),
		items:   make(map[sources.SourceName][]queueItem),
		notify:  make(map[sources.SourceName]chan struct{}),
	}
}

func (q *queue) enqueue(source sources.SourceName, chapterList []chapters.Chapter) int {
	sorted := slices.Clone(chapterList)
	slices.SortFunc(sorted, func(a, b chapters.Chapter) int {
		switch {
		case a.Number < b.Number:
			return -1
		case a.Number > b.Number:
			return 1
		default:
			return 0
		}
	})

	q.mu.Lock()
	defer q.mu.Unlock()

	added := 0
	for _, chapter := range sorted {
		if _, ok := q.pending[chapter.ID]; ok {
			continue
		}

		q.pending[chapter.ID] = struct{}{}
		q.items[source] = append(q.items[source], queueItem{
			ChapterID: chapter.ID,
			Number:    chapter.Number,
		})
		added++
	}

	if added > 0 {
		q.signalLocked(source)
	}

	return added
}

func (q *queue) pop(ctx context.Context, source sources.SourceName) (queueItem, bool) {
	for {
		q.mu.Lock()
		items := q.items[source]
		if len(items) > 0 {
			item := items[0]
			q.items[source] = items[1:]
			q.mu.Unlock()

			return item, true
		}

		notify := q.notifyChannelLocked(source)
		q.mu.Unlock()

		select {
		case <-notify:
		case <-ctx.Done():
			return queueItem{}, false
		}
	}
}

func (q *queue) done(chapterID uuid.UUID) {
	q.mu.Lock()
	delete(q.pending, chapterID)
	q.mu.Unlock()
}

func (q *queue) removeComicChapters(chapterIDs map[uuid.UUID]struct{}) {
	q.mu.Lock()
	defer q.mu.Unlock()

	for source, items := range q.items {
		filtered := items[:0]
		for _, item := range items {
			if _, belongs := chapterIDs[item.ChapterID]; belongs {
				delete(q.pending, item.ChapterID)

				continue
			}

			filtered = append(filtered, item)
		}

		if len(filtered) == 0 {
			delete(q.items, source)
		} else {
			q.items[source] = filtered
		}
	}
}

func (q *queue) pendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()

	return len(q.pending)
}

func (q *queue) notifyChannelLocked(source sources.SourceName) chan struct{} {
	if ch, ok := q.notify[source]; ok {
		return ch
	}

	ch := make(chan struct{}, 1)
	q.notify[source] = ch

	return ch
}

func (q *queue) signalLocked(source sources.SourceName) {
	ch := q.notifyChannelLocked(source)

	select {
	case ch <- struct{}{}:
	default:
	}
}

// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/chapters"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

func TestQueueEnqueueIsIdempotent(t *testing.T) {
	t.Parallel()

	q := newQueue()
	chapterID := uuid.New()

	added := q.enqueue(sources.SourceAsuraScans, []chapters.Chapter{
		{ID: chapterID, Number: 2},
		{ID: chapterID, Number: 2},
	})
	if added != 1 {
		t.Fatalf("added = %d, want 1", added)
	}

	added = q.enqueue(sources.SourceAsuraScans, []chapters.Chapter{{ID: chapterID, Number: 2}})
	if added != 0 {
		t.Fatalf("second enqueue added = %d, want 0", added)
	}

	if q.pendingCount() != 1 {
		t.Fatalf("pending = %d, want 1", q.pendingCount())
	}
}

func TestQueueOrdersChaptersByNumber(t *testing.T) {
	t.Parallel()

	q := newQueue()
	firstID := uuid.New()
	secondID := uuid.New()

	q.enqueue(sources.SourceAsuraScans, []chapters.Chapter{
		{ID: secondID, Number: 2},
		{ID: firstID, Number: 1},
	})

	item, ok := q.popNoWait(sources.SourceAsuraScans)
	if !ok || item.ChapterID != firstID {
		t.Fatalf("first item = %+v, ok = %v", item, ok)
	}

	item, ok = q.popNoWait(sources.SourceAsuraScans)
	if !ok || item.ChapterID != secondID {
		t.Fatalf("second item = %+v, ok = %v", item, ok)
	}
}

func TestQueueRemoveComicChapters(t *testing.T) {
	t.Parallel()

	q := newQueue()
	firstID := uuid.New()
	secondID := uuid.New()
	otherComicChapterID := uuid.New()

	q.enqueue(sources.SourceAsuraScans, []chapters.Chapter{
		{ID: firstID, Number: 1},
		{ID: secondID, Number: 2},
		{ID: otherComicChapterID, Number: 3},
	})

	q.removeComicChapters(map[uuid.UUID]struct{}{
		firstID:  {},
		secondID: {},
	})

	if q.pendingCount() != 1 {
		t.Fatalf("pending = %d, want 1", q.pendingCount())
	}

	item, ok := q.popNoWait(sources.SourceAsuraScans)
	if !ok || item.ChapterID != otherComicChapterID {
		t.Fatalf("remaining item = %+v, ok = %v", item, ok)
	}
}

func (q *queue) popNoWait(source sources.SourceName) (queueItem, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	items := q.items[source]
	if len(items) == 0 {
		return queueItem{}, false
	}

	item := items[0]
	q.items[source] = items[1:]

	return item, true
}

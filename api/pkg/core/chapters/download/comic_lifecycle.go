// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import (
	"context"

	"github.com/google/uuid"
)

type comicTracker struct {
	ctx    context.Context
	cancel context.CancelFunc
	active int
}

func (w *Worker) beginComic(parent context.Context, comicID uuid.UUID) (context.Context, func()) {
	w.comicMu.Lock()

	tracker, ok := w.comicTrackers[comicID]
	if !ok {
		comicCtx, cancel := context.WithCancel(parent)
		tracker = &comicTracker{
			ctx:    comicCtx,
			cancel: cancel,
		}
		w.comicTrackers[comicID] = tracker
	}

	tracker.active++
	comicCtx := tracker.ctx
	w.comicMu.Unlock()

	return comicCtx, func() {
		w.comicMu.Lock()
		defer w.comicMu.Unlock()

		tracker, ok := w.comicTrackers[comicID]
		if !ok {
			return
		}

		tracker.active--
		if tracker.active == 0 {
			delete(w.comicTrackers, comicID)
		}

		w.comicCond.Broadcast()
	}
}

func (w *Worker) cancelComicWork(comicID uuid.UUID) {
	w.comicMu.Lock()
	defer w.comicMu.Unlock()

	tracker, ok := w.comicTrackers[comicID]
	if !ok {
		return
	}

	tracker.cancel()
}

func (w *Worker) waitComicWorkDone(comicID uuid.UUID) {
	w.comicMu.Lock()
	defer w.comicMu.Unlock()

	for {
		tracker, ok := w.comicTrackers[comicID]
		if !ok || tracker.active == 0 {
			return
		}

		w.comicCond.Wait()
	}
}

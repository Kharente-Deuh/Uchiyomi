// SPDX-License-Identifier: AGPL-3.0-or-later

package utils_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

func TestLoopStopsOnContextCancel(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := utils.Loop(ctx, utils.LoopOpts{
		Interval: time.Hour,
		Fn: func(context.Context) error {
			t.Error("fn ne doit pas être appelée quand le contexte est déjà annulé")

			return nil
		},
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("utils.Loop() = %v, want context.Canceled", err)
	}
}

func TestLoopStopsOnDeadline(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()

	err := utils.Loop(ctx, utils.LoopOpts{
		Interval: time.Hour,
		Fn:       func(context.Context) error { return nil },
	})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("utils.Loop() = %v, want context.DeadlineExceeded", err)
	}
}

func TestLoopRunsFnBeforeFirstWait(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	base, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()

	ctx, cancel := context.WithCancel(base)
	defer cancel()

	err := utils.Loop(ctx, utils.LoopOpts{
		Interval: time.Hour,
		Fn: func(context.Context) error {
			calls.Add(1)
			cancel()

			return nil
		},
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("utils.Loop() = %v, want context.Canceled", err)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("fn appelée %d fois, want 1 avant la première attente", got)
	}
}

func TestLoopCallsFnRepeatedly(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	_ = utils.Loop(ctx, utils.LoopOpts{
		Interval: 10 * time.Millisecond,
		Fn: func(context.Context) error {
			calls.Add(1)

			return nil
		},
	})

	if got := calls.Load(); got < 2 {
		t.Errorf("fn appelée %d fois en 120ms avec un intervalle de 10ms, want >= 2", got)
	}
}

func TestLoopWaitsIntervalAfterFnReturns(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	_ = utils.Loop(ctx, utils.LoopOpts{
		Interval: 40 * time.Millisecond,
		Fn: func(context.Context) error {
			calls.Add(1)
			time.Sleep(40 * time.Millisecond)

			return nil
		},
	})

	if got := calls.Load(); got > 2 {
		t.Errorf("fn appelée %d fois en 120ms, want <= 2 : l'attente ne repart pas de la fin de fn", got)
	}
}

func TestLoopReturnsFnError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("boom")

	var calls atomic.Int32

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := utils.Loop(ctx, utils.LoopOpts{
		Interval: time.Millisecond,
		Fn: func(context.Context) error {
			calls.Add(1)

			return sentinel
		},
	})

	if !errors.Is(err, sentinel) {
		t.Errorf("utils.Loop() = %v, want %v", err, sentinel)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("fn appelée %d fois, want 1 (la boucle doit s'arrêter à la première erreur)", got)
	}
}

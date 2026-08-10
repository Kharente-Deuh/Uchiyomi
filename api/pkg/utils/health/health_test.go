// SPDX-License-Identifier: AGPL-3.0-or-later

package health_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/health"
)

func TestNewRegistryStartsLatchesAsStarting(t *testing.T) {
	reg := health.NewRegistry("migrations", "asura")

	rep := reg.Snapshot(context.Background())

	if rep.Status != health.StatusStarting {
		t.Fatalf("overall status = %q, want %q", rep.Status, health.StatusStarting)
	}

	for _, name := range []string{"migrations", "asura"} {
		if got := rep.Components[name].Status; got != health.StatusStarting {
			t.Errorf("%s = %q, want %q", name, got, health.StatusStarting)
		}
	}
}

func TestSetNilMarksLatchOK(t *testing.T) {
	reg := health.NewRegistry("migrations")

	reg.Set("migrations", nil)

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusOK {
		t.Fatalf("overall status = %q, want %q", rep.Status, health.StatusOK)
	}

	if rep.Components["migrations"].Reason != "" {
		t.Errorf("reason = %q, want empty", rep.Components["migrations"].Reason)
	}
}

func TestSetErrorMarksLatchFailedWithReason(t *testing.T) {
	reg := health.NewRegistry("migrations")

	reg.Set("migrations", errors.New("unknown column"))

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusFailed {
		t.Fatalf("overall status = %q, want %q", rep.Status, health.StatusFailed)
	}

	if got := rep.Components["migrations"].Reason; got != "unknown column" {
		t.Errorf("reason = %q, want %q", got, "unknown column")
	}
}

func TestFailedDominatesStarting(t *testing.T) {
	reg := health.NewRegistry("migrations", "asura")

	reg.Set("migrations", errors.New("boom"))

	if got := reg.Snapshot(context.Background()).Status; got != health.StatusFailed {
		t.Fatalf("overall status = %q, want %q", got, health.StatusFailed)
	}
}

func TestProbeSuccessIsOK(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return nil })

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusOK {
		t.Fatalf("overall status = %q, want %q", rep.Status, health.StatusOK)
	}

	if got := rep.Components["db"].Status; got != health.StatusOK {
		t.Errorf("db = %q, want %q", got, health.StatusOK)
	}
}

func TestProbeErrorIsFailedWithReason(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return errors.New("connection refused") })

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusFailed {
		t.Fatalf("overall status = %q, want %q", rep.Status, health.StatusFailed)
	}

	if got := rep.Components["db"].Reason; got != "connection refused" {
		t.Errorf("reason = %q, want %q", got, "connection refused")
	}
}

func TestProbeReceivesCallerContext(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(ctx context.Context) error { return ctx.Err() })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if got := reg.Snapshot(ctx).Components["db"].Status; got != health.StatusFailed {
		t.Fatalf("db = %q, want %q", got, health.StatusFailed)
	}
}

func TestProbeIsReevaluatedOnEachSnapshot(t *testing.T) {
	var calls int
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error {
		calls++

		return nil
	})

	reg.Snapshot(context.Background())
	reg.Snapshot(context.Background())

	if calls != 2 {
		t.Fatalf("probe called %d times, want 2", calls)
	}
}

func TestSetOnUnknownNameCreatesComponent(t *testing.T) {
	reg := health.NewRegistry()

	reg.Set("tardif", nil)

	if got := reg.Snapshot(context.Background()).Components["tardif"].Status; got != health.StatusOK {
		t.Fatalf("late = %q, want %q", got, health.StatusOK)
	}
}

func TestSnapshotMarksProbesAndNotLatches(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.AddProbe("db", func(context.Context) error { return nil })

	rep := reg.Snapshot(context.Background())

	if !rep.Components["db"].Probe {
		t.Error("db is not marked as probe")
	}

	if rep.Components["migrations"].Probe {
		t.Error("migrations is marked as probe")
	}
}

func TestLatchStatusReturnsLatchWithoutRunningProbes(t *testing.T) {
	var calls int

	reg := health.NewRegistry("migrations")
	reg.AddProbe("db", func(context.Context) error {
		calls++

		return nil
	})

	if got := reg.LatchStatus("migrations"); got != health.StatusStarting {
		t.Fatalf("migrations = %q, want %q", got, health.StatusStarting)
	}

	reg.Set("migrations", nil)

	if got := reg.LatchStatus("migrations"); got != health.StatusOK {
		t.Fatalf("migrations = %q, want %q", got, health.StatusOK)
	}

	if calls != 0 {
		t.Errorf("probe called %d times, want 0", calls)
	}
}

func TestLatchStatusOnUnknownNameIsStarting(t *testing.T) {
	reg := health.NewRegistry("migrations")

	if got := reg.LatchStatus("migratons"); got != health.StatusStarting {
		t.Fatalf("unknown name = %q, want %q", got, health.StatusStarting)
	}
}

func TestLatchStatusIgnoresProbes(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return nil })

	if got := reg.LatchStatus("db"); got != health.StatusStarting {
		t.Fatalf("db = %q, want %q", got, health.StatusStarting)
	}
}

func TestConcurrentSetAndSnapshot(t *testing.T) {
	reg := health.NewRegistry("a", "b")
	reg.AddProbe("db", func(context.Context) error { return nil })

	var wg sync.WaitGroup

	for i := range 50 {
		wg.Add(3)

		go func() {
			defer wg.Done()

			reg.LatchStatus("a")
		}()

		go func() {
			defer wg.Done()

			reg.Set("a", nil)
		}()

		go func() {
			defer wg.Done()

			if i%2 == 0 {
				reg.Set("b", errors.New("boom"))
			}

			reg.Snapshot(context.Background())
		}()
	}

	wg.Wait()
}

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
		t.Fatalf("statut global = %q, attendu %q", rep.Status, health.StatusStarting)
	}

	for _, name := range []string{"migrations", "asura"} {
		if got := rep.Components[name].Status; got != health.StatusStarting {
			t.Errorf("%s = %q, attendu %q", name, got, health.StatusStarting)
		}
	}
}

func TestSetNilMarksLatchOK(t *testing.T) {
	reg := health.NewRegistry("migrations")

	reg.Set("migrations", nil)

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusOK {
		t.Fatalf("statut global = %q, attendu %q", rep.Status, health.StatusOK)
	}

	if rep.Components["migrations"].Reason != "" {
		t.Errorf("raison = %q, attendu vide", rep.Components["migrations"].Reason)
	}
}

func TestSetErrorMarksLatchFailedWithReason(t *testing.T) {
	reg := health.NewRegistry("migrations")

	reg.Set("migrations", errors.New("colonne inconnue"))

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusFailed {
		t.Fatalf("statut global = %q, attendu %q", rep.Status, health.StatusFailed)
	}

	if got := rep.Components["migrations"].Reason; got != "colonne inconnue" {
		t.Errorf("raison = %q, attendu %q", got, "colonne inconnue")
	}
}

func TestFailedDominatesStarting(t *testing.T) {
	reg := health.NewRegistry("migrations", "asura")

	reg.Set("migrations", errors.New("boom"))

	if got := reg.Snapshot(context.Background()).Status; got != health.StatusFailed {
		t.Fatalf("statut global = %q, attendu %q", got, health.StatusFailed)
	}
}

func TestProbeSuccessIsOK(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return nil })

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusOK {
		t.Fatalf("statut global = %q, attendu %q", rep.Status, health.StatusOK)
	}

	if got := rep.Components["db"].Status; got != health.StatusOK {
		t.Errorf("db = %q, attendu %q", got, health.StatusOK)
	}
}

func TestProbeErrorIsFailedWithReason(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return errors.New("connexion refusée") })

	rep := reg.Snapshot(context.Background())
	if rep.Status != health.StatusFailed {
		t.Fatalf("statut global = %q, attendu %q", rep.Status, health.StatusFailed)
	}

	if got := rep.Components["db"].Reason; got != "connexion refusée" {
		t.Errorf("raison = %q, attendu %q", got, "connexion refusée")
	}
}

func TestProbeReceivesCallerContext(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(ctx context.Context) error { return ctx.Err() })

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if got := reg.Snapshot(ctx).Components["db"].Status; got != health.StatusFailed {
		t.Fatalf("db = %q, attendu %q", got, health.StatusFailed)
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
		t.Fatalf("sonde appelée %d fois, attendu 2", calls)
	}
}

func TestSetOnUnknownNameCreatesComponent(t *testing.T) {
	reg := health.NewRegistry()

	reg.Set("tardif", nil)

	if got := reg.Snapshot(context.Background()).Components["tardif"].Status; got != health.StatusOK {
		t.Fatalf("tardif = %q, attendu %q", got, health.StatusOK)
	}
}

func TestSnapshotMarksProbesAndNotLatches(t *testing.T) {
	reg := health.NewRegistry("migrations")
	reg.AddProbe("db", func(context.Context) error { return nil })

	rep := reg.Snapshot(context.Background())

	if !rep.Components["db"].Probe {
		t.Error("db n'est pas marqué comme sonde")
	}

	if rep.Components["migrations"].Probe {
		t.Error("migrations est marqué comme sonde")
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
		t.Fatalf("migrations = %q, attendu %q", got, health.StatusStarting)
	}

	reg.Set("migrations", nil)

	if got := reg.LatchStatus("migrations"); got != health.StatusOK {
		t.Fatalf("migrations = %q, attendu %q", got, health.StatusOK)
	}

	if calls != 0 {
		t.Errorf("sonde appelée %d fois, attendu 0", calls)
	}
}

func TestLatchStatusOnUnknownNameIsStarting(t *testing.T) {
	reg := health.NewRegistry("migrations")

	if got := reg.LatchStatus("migratons"); got != health.StatusStarting {
		t.Fatalf("nom inconnu = %q, attendu %q", got, health.StatusStarting)
	}
}

func TestLatchStatusIgnoresProbes(t *testing.T) {
	reg := health.NewRegistry()
	reg.AddProbe("db", func(context.Context) error { return nil })

	if got := reg.LatchStatus("db"); got != health.StatusStarting {
		t.Fatalf("db = %q, attendu %q", got, health.StatusStarting)
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

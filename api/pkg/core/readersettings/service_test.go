// SPDX-License-Identifier: AGPL-3.0-or-later

package readersettings_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/readersettings"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

//nolint:govet // fieldalignment on a test fake is not worth the unreadable field order
type fakeRepo struct {
	listErr     error
	upsertErr   error
	listedFor   uuid.UUID
	stored      []readersettings.Profile
	lastUpsert  readersettings.UpsertOpts
	listCalls   int
	upsertCalls int
}

func (f *fakeRepo) ListByUser(_ context.Context, userID uuid.UUID) ([]readersettings.Profile, error) {
	f.listCalls++
	f.listedFor = userID

	return f.stored, f.listErr
}

func (f *fakeRepo) Upsert(_ context.Context, opts readersettings.UpsertOpts) (readersettings.Profile, error) {
	f.upsertCalls++
	f.lastUpsert = opts

	return readersettings.Profile{
		Type:        opts.Type,
		ReadingMode: opts.ReadingMode,
		PageScale:   opts.PageScale,
		DoublePage:  opts.DoublePage,
	}, f.upsertErr
}

func newService(t *testing.T, repo readersettings.Repository) *readersettings.Service {
	t.Helper()

	svc, err := readersettings.NewService(readersettings.Deps{Repository: repo})
	if err != nil {
		t.Fatalf("NewService: %v", err)
	}

	return svc
}

func TestNewServiceValidatesDeps(t *testing.T) {
	t.Parallel()

	_, err := readersettings.NewService(readersettings.Deps{})
	if err == nil {
		t.Fatal("NewService with empty deps must fail")
	}
}

func TestListForUserReturnsFactoryDefaultsWithoutUpsert(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo)
	userID := uuid.New()

	got, err := svc.ListForUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}

	if repo.listCalls != 1 || repo.listedFor != userID {
		t.Errorf("ListByUser calls=%d user=%s", repo.listCalls, repo.listedFor)
	}

	if repo.upsertCalls != 0 {
		t.Fatalf("Upsert called %d times, want 0", repo.upsertCalls)
	}

	want := readersettings.AllTypes()
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d", len(got), len(want))
	}

	for i, typ := range want {
		if got[i] != readersettings.DefaultProfile(typ) {
			t.Errorf("items[%d] = %+v, want default %q", i, got[i], typ)
		}
	}
}

func TestListForUserOverlaysStoredMangaOnly(t *testing.T) {
	t.Parallel()

	saved := readersettings.Profile{
		Type:        sources.SeriesTypeManga,
		ReadingMode: readersettings.ReadingModePagedLTR,
		PageScale:   readersettings.PageScaleFitHeight,
		DoublePage:  true,
	}
	repo := &fakeRepo{stored: []readersettings.Profile{saved}}
	svc := newService(t, repo)

	got, err := svc.ListForUser(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("ListForUser: %v", err)
	}

	if got[0] != saved {
		t.Errorf("manga = %+v, want stored", got[0])
	}

	if got[1] != readersettings.DefaultProfile(sources.SeriesTypeManhua) {
		t.Errorf("manhua changed: %+v", got[1])
	}

	if got[2] != readersettings.DefaultProfile(sources.SeriesTypeManhwa) {
		t.Errorf("manhwa changed: %+v", got[2])
	}
}

func TestListForUserWrapsRepoError(t *testing.T) {
	t.Parallel()

	cause := errors.New("list failed")
	repo := &fakeRepo{listErr: cause}
	svc := newService(t, repo)

	_, err := svc.ListForUser(context.Background(), uuid.New())
	if !errors.Is(err, cause) {
		t.Errorf("ListForUser = %v, original error no longer reachable", err)
	}
}

func TestReplaceMangaDoesNotWriteManhwa(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo)
	userID := uuid.New()

	got, err := svc.Replace(context.Background(), readersettings.ReplaceOpts{
		UserID:      userID,
		Type:        sources.SeriesTypeManga,
		ReadingMode: readersettings.ReadingModePagedRTL,
		PageScale:   readersettings.PageScaleFitHeight,
		DoublePage:  true,
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	if repo.upsertCalls != 1 {
		t.Fatalf("Upsert calls = %d, want 1", repo.upsertCalls)
	}

	if repo.lastUpsert.UserID != userID || repo.lastUpsert.Type != sources.SeriesTypeManga {
		t.Errorf("upsert = %+v", repo.lastUpsert)
	}

	if repo.lastUpsert.Type == sources.SeriesTypeManhwa {
		t.Fatal("upsert wrote manhwa")
	}

	if got.Type != sources.SeriesTypeManga || !got.DoublePage {
		t.Errorf("Replace() = %+v", got)
	}
}

func TestReplaceRejectsWebtoonDoublePage(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo)

	_, err := svc.Replace(context.Background(), readersettings.ReplaceOpts{
		UserID:      uuid.New(),
		Type:        sources.SeriesTypeManhwa,
		ReadingMode: readersettings.ReadingModeWebtoon,
		PageScale:   readersettings.PageScaleFitWidth,
		DoublePage:  true,
	})
	if !errors.Is(err, readersettings.ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}

	if repo.upsertCalls != 0 {
		t.Errorf("Upsert called on invalid combo")
	}
}

func TestReplaceRejectsUnknownType(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo)

	_, err := svc.Replace(context.Background(), readersettings.ReplaceOpts{
		UserID:      uuid.New(),
		Type:        sources.SeriesType("webcomic"),
		ReadingMode: readersettings.ReadingModePagedLTR,
		PageScale:   readersettings.PageScaleFitScreen,
	})
	if !errors.Is(err, readersettings.ErrInvalid) {
		t.Errorf("err = %v, want ErrInvalid", err)
	}

	if repo.upsertCalls != 0 {
		t.Error("Upsert called for unknown type")
	}
}

func TestReplacePassesCallerUserID(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	svc := newService(t, repo)
	alice := uuid.New()

	_, err := svc.Replace(context.Background(), readersettings.ReplaceOpts{
		UserID:      alice,
		Type:        sources.SeriesTypeManhua,
		ReadingMode: readersettings.ReadingModePagedLTR,
		PageScale:   readersettings.PageScaleFitScreen,
	})
	if err != nil {
		t.Fatalf("Replace: %v", err)
	}

	if repo.lastUpsert.UserID != alice {
		t.Errorf("UserID = %s, want alice", repo.lastUpsert.UserID)
	}
}

func TestReplaceWrapsUpsertError(t *testing.T) {
	t.Parallel()

	cause := errors.New("upsert failed")
	repo := &fakeRepo{upsertErr: cause}
	svc := newService(t, repo)

	_, err := svc.Replace(context.Background(), readersettings.ReplaceOpts{
		UserID:      uuid.New(),
		Type:        sources.SeriesTypeManga,
		ReadingMode: readersettings.ReadingModePagedRTL,
		PageScale:   readersettings.PageScaleFitScreen,
	})
	if !errors.Is(err, cause) {
		t.Errorf("Replace = %v, original error no longer reachable", err)
	}
}

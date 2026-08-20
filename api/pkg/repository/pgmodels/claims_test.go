// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
	"github.com/kharente-deuh/uchiyomi-server/pkg/sources"
)

const (
	testEmail      = "bob@example.com"
	testEmailClaim = "email"
)

func TestClaimsValue(t *testing.T) {
	t.Parallel()

	t.Run("nil serializes empty object", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims

		got, err := c.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		b, ok := got.([]byte)
		if !ok {
			t.Fatalf("Value() = %T, want []byte", got)
		}

		if string(b) != "{}" {
			t.Errorf("Value() = %q, want %q", b, "{}")
		}
	})

	t.Run("map serialized as JSON", func(t *testing.T) {
		t.Parallel()

		got, err := pgmodels.Claims{testEmailClaim: testEmail}.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if want := `{"email":"bob@example.com"}`; string(got.([]byte)) != want {
			t.Errorf("Value() = %s, want %s", got, want)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		t.Parallel()

		got, err := pgmodels.Claims{}.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if string(got.([]byte)) != "{}" {
			t.Errorf("Value() = %s, want {}", got)
		}
	})
}

func TestClaimsScan(t *testing.T) {
	t.Parallel()

	t.Run("nil resets map to nil", func(t *testing.T) {
		t.Parallel()

		c := pgmodels.Claims{"stale": true}
		if err := c.Scan(nil); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if c != nil {
			t.Errorf("Scan(nil) = %v, want nil", c)
		}
	})

	t.Run("JSON valide", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims
		if err := c.Scan([]byte(`{"email":"bob@example.com","admin":true}`)); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		want := pgmodels.Claims{testEmailClaim: testEmail, "admin": true}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("Scan() = %v, want %v", c, want)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims
		if err := c.Scan([]byte(`{`)); err == nil {
			t.Error("Scan on truncated JSON = nil, want error")
		}
	})

	t.Run("type inattendu", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims

		err := c.Scan(`{"email":"bob@example.com"}`)
		if err == nil {
			t.Fatal("Scan(string) = nil, want error")
		}

		if want := "claims: type string inattendu"; err.Error() != want {
			t.Errorf("err = %q, want %q", err.Error(), want)
		}
	})
}

func TestClaimsRoundTrip(t *testing.T) {
	t.Parallel()

	original := pgmodels.Claims{testEmailClaim: testEmail, "groups": []any{"admins", "users"}}

	encoded, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	var decoded pgmodels.Claims
	if err := decoded.Scan(encoded); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("aller-retour = %v, want %v", decoded, original)
	}
}

func TestGormDataType(t *testing.T) {
	t.Parallel()

	if got := (pgmodels.Claims{}).GormDataType(); got != "jsonb" {
		t.Errorf("GormDataType() = %q, want %q", got, "jsonb")
	}
}

func TestTableNames(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"user":               pgmodels.User{}.TableName(),
		"password creds":     pgmodels.PasswordCreds{}.TableName(),
		"oidc provider":      pgmodels.OIDCProvider{}.TableName(),
		"federated identity": pgmodels.FederatedIdentity{}.TableName(),
		"session":            pgmodels.Session{}.TableName(),
		"comic":              pgmodels.Comic{}.TableName(),
		"chapter":            pgmodels.Chapter{}.TableName(),
		"library entry":      pgmodels.LibraryEntry{}.TableName(),
		"reader settings":    pgmodels.ReaderSettings{}.TableName(),
		"reading progress":   pgmodels.ReadingProgress{}.TableName(),
	}

	want := map[string]string{
		"user":               "users",
		"password creds":     "password_credentials",
		"oidc provider":      "oidc_providers",
		"federated identity": "federated_identities",
		"session":            "sessions",
		"comic":              "comics",
		"chapter":            "chapters",
		"library entry":      "library_entries",
		"reader settings":    "reader_settings",
		"reading progress":   "reading_progress",
	}

	for model, got := range tests {
		if got != want[model] {
			t.Errorf("%s: TableName() = %q, want %q", model, got, want[model])
		}
	}
}

func TestReadingProgressConstraints(t *testing.T) {
	t.Parallel()

	typ := reflect.TypeOf(pgmodels.ReadingProgress{})

	for _, name := range []string{"LibraryEntry", "Chapter"} {
		field, ok := typ.FieldByName(name)
		if !ok {
			t.Fatalf("%s field not found", name)
		}

		tag := field.Tag.Get("gorm")
		if !strings.Contains(tag, "constraint:OnDelete:CASCADE") {
			t.Errorf("%s gorm tag = %q, want constraint:OnDelete:CASCADE", name, tag)
		}
	}
}

func TestChapterDomain(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	comicID := uuid.New()
	publishedAt := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	early := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC)
	model := pgmodels.Chapter{
		ID:                id,
		ComicID:           comicID,
		SourceChapterSlug: "chapter-1",
		Number:            1.5,
		Title:             "Chapter 1",
		PagesNb:           20,
		PublishedAt:       publishedAt,
		EarlyAccessUntil:  &early,
		Download:          40,
	}

	got := model.Domain()
	if got.ID != id || got.ComicID != comicID || got.SourceChapterSlug != "chapter-1" {
		t.Errorf("Domain() ids = %+v", got)
	}

	if got.Number != 1.5 || got.PagesNb != 20 || got.Download != 40 {
		t.Errorf("Domain() numbers = %+v", got)
	}

	if !got.PublishedAt.Equal(publishedAt) || got.EarlyAccessUntil == nil || !got.EarlyAccessUntil.Equal(early) {
		t.Errorf("Domain() times = %+v", got)
	}
}

func TestComicDomainAndEnums(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	created := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	comicType := pgmodels.ComicTypeManhwa
	status := pgmodels.ComicStatusOngoing
	model := pgmodels.Comic{
		ID:           id,
		Source:       sources.SourceAsuraScans,
		Slug:         "solo-leveling",
		Title:        "Solo Leveling",
		Status:       status,
		ComicType:    comicType,
		ChapterCount: 200,
		CreatedAt:    created,
		UpdatedAt:    created,
	}

	got := model.Domain()
	if got.ID != id || got.Slug != "solo-leveling" || got.Type != sources.SeriesTypeManhwa {
		t.Errorf("Domain() = %+v", got)
	}

	if got.Status != sources.SeriesStatusOngoing {
		t.Errorf("Status = %q, want ongoing", got.Status)
	}

	if pgmodels.ComicTypeFromDomain(sources.SeriesTypeManga) != pgmodels.ComicTypeManga {
		t.Error("ComicTypeFromDomain(manga) mismatch")
	}

	if pgmodels.ComicStatusFromDomain(sources.SeriesStatusHiatus) != pgmodels.ComicStatusHiatus {
		t.Error("ComicStatusFromDomain(hiatus) mismatch")
	}
}

func TestLibraryEntryDomain(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	userID := uuid.New()
	comicID := uuid.New()
	addedAt := time.Date(2024, 5, 1, 0, 0, 0, 0, time.UTC)
	model := pgmodels.LibraryEntry{ID: id, UserID: userID, ComicID: comicID, AddedAt: addedAt}

	got := model.Domain()
	if got.ID != id || got.UserID != userID || got.ComicID != comicID || !got.AddedAt.Equal(addedAt) {
		t.Errorf("Domain() = %+v", got)
	}
}

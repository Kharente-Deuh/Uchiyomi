// SPDX-License-Identifier: AGPL-3.0-or-later

package sessions_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/users"
)

const (
	day      = 24 * time.Hour
	idlePwd  = 30 * day
	absPwd   = 90 * day
	idleOIDC = day
	absOIDC  = 30 * day
	renew    = time.Hour
)

func validConfig() sessions.ServiceConfig {
	return sessions.ServiceConfig{
		Password:       sessions.TTL{Idle: idlePwd, Absolute: absPwd},
		OIDC:           sessions.TTL{Idle: idleOIDC, Absolute: absOIDC},
		RenewThreshold: renew,
	}
}

type fakeRepository struct {
	session      *sessions.Session
	user         *users.User
	gotHash      []byte
	gotExpiry    time.Time
	insertErr    error
	getErr       error
	updateErr    error
	deleteErr    error
	gotInsert    sessions.InsertSessionOpts
	inserts      int
	gets         int
	updates      int
	deleteHashes int
	deleteUsers  int
	gotUserID    uuid.UUID
	gotID        uuid.UUID
}

func (f *fakeRepository) Insert(_ context.Context, opts sessions.InsertSessionOpts) (*sessions.Session, error) {
	f.inserts++
	f.gotInsert = opts

	if f.insertErr != nil {
		return nil, f.insertErr
	}

	return &sessions.Session{
		ID:         uuid.New(),
		UserID:     opts.UserID,
		AuthMethod: opts.AuthMethod,
		ExpiresAt:  opts.ExpiresAt,
		CreatedAt:  opts.ExpiresAt.Add(-time.Minute),
	}, nil
}

func (f *fakeRepository) GetByTokenHash(_ context.Context, hash []byte) (*sessions.Session, *users.User, error) {
	f.gets++
	f.gotHash = hash

	if f.getErr != nil {
		return nil, nil, f.getErr
	}

	return f.session, f.user, nil
}

func (f *fakeRepository) UpdateExpiry(_ context.Context, id uuid.UUID, expiresAt time.Time) error {
	f.updates++
	f.gotID = id
	f.gotExpiry = expiresAt

	return f.updateErr
}

func (f *fakeRepository) DeleteByTokenHash(_ context.Context, hash []byte) error {
	f.deleteHashes++
	f.gotHash = hash

	return f.deleteErr
}

func (f *fakeRepository) DeleteByUserID(_ context.Context, id uuid.UUID) error {
	f.deleteUsers++
	f.gotUserID = id

	return f.deleteErr
}

func (f *fakeRepository) DeleteExpired(context.Context, time.Time) (int64, error) {
	panic("DeleteExpired ne doit pas être appelée par le service")
}

func TestNewRejectsInvalidConfig(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		mutate  func(*sessions.ServiceConfig)
		wantErr string
	}{
		"idle mot de passe nul": {
			mutate:  func(c *sessions.ServiceConfig) { c.Password.Idle = 0 },
			wantErr: "cfg.Validate: password ttl: idle must be positive",
		},
		"absolu inférieur à idle": {
			mutate:  func(c *sessions.ServiceConfig) { c.OIDC.Absolute = time.Minute },
			wantErr: "cfg.Validate: oidc ttl: absolute must not be lower than idle",
		},
		"seuil de prolongation nul": {
			mutate:  func(c *sessions.ServiceConfig) { c.RenewThreshold = 0 },
			wantErr: "cfg.Validate: renewThreshold must be positive",
		},
		"seuil au-delà du plus petit idle": {
			mutate:  func(c *sessions.ServiceConfig) { c.RenewThreshold = 2 * day },
			wantErr: "cfg.Validate: renewThreshold must be lower than the shortest idle ttl",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			cfg := validConfig()
			tc.mutate(&cfg)

			svc, err := sessions.NewService(cfg, sessions.ServiceDeps{Repository: &fakeRepository{}})
			if err == nil {
				t.Fatalf("New() = nil, want %q", tc.wantErr)
			}

			if svc != nil {
				t.Error("New a renvoyé un service en plus de l'erreur")
			}

			if err.Error() != tc.wantErr {
				t.Errorf("New() = %q, want %q", err.Error(), tc.wantErr)
			}
		})
	}
}

func TestNewRequiresRepository(t *testing.T) {
	t.Parallel()

	svc, err := sessions.NewService(validConfig(), sessions.ServiceDeps{})
	if err == nil {
		t.Fatal("New sans repository doit échouer")
	}

	if svc != nil {
		t.Error("New a renvoyé un service en plus de l'erreur")
	}

	if want := "deps.Validate: repository is required"; err.Error() != want {
		t.Errorf("New() = %q, want %q", err.Error(), want)
	}
}

func TestNewAcceptsValidConfig(t *testing.T) {
	t.Parallel()

	svc, err := sessions.NewService(validConfig(), sessions.ServiceDeps{Repository: &fakeRepository{}})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if svc == nil {
		t.Error("New a renvoyé un service nil sans erreur")
	}
}

func frozenSvc(t *testing.T, repo *fakeRepository, now time.Time) *sessions.Service {
	t.Helper()

	svc, err := sessions.NewService(validConfig(), sessions.ServiceDeps{
		Repository: repo,
		Now:        func() time.Time { return now },
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return svc
}

func TestCreateIssuesA256BitToken(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc := frozenSvc(t, repo, time.Now())

	first, err := svc.Create(context.Background(), sessions.CreateSessionOpts{
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if len(first.Token) != 43 {
		t.Errorf("len(Token) = %d, want 43 (32 octets en base64url sans padding)", len(first.Token))
	}

	raw, err := base64.RawURLEncoding.DecodeString(first.Token)
	if err != nil {
		t.Fatalf("le token n'est pas du base64url: %v", err)
	}

	if len(raw) != 32 {
		t.Errorf("le token décode %d octets, want 32", len(raw))
	}

	second, err := svc.Create(context.Background(), sessions.CreateSessionOpts{
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if first.Token == second.Token {
		t.Error("deux appels ont produit le même token")
	}
}

func TestCreateStoresHashNeverToken(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	svc := frozenSvc(t, repo, time.Now())

	issued, err := svc.Create(context.Background(), sessions.CreateSessionOpts{
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	want := sha256.Sum256([]byte(issued.Token))
	if !bytes.Equal(repo.gotInsert.TokenHash, want[:]) {
		t.Errorf("TokenHash = %x, want %x", repo.gotInsert.TokenHash, want)
	}

	if len(repo.gotInsert.TokenHash) != sha256.Size {
		t.Errorf("len(TokenHash) = %d, want %d", len(repo.gotInsert.TokenHash), sha256.Size)
	}

	if string(repo.gotInsert.TokenHash) == issued.Token {
		t.Error("le token en clair est parti au repository")
	}
}

func TestCreateAppliesTTLOfAuthMethod(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		method sessions.AuthMethod
		want   time.Duration
	}{
		"mot de passe": {method: sessions.AuthMethodPassword, want: idlePwd},
		"oidc":         {method: sessions.AuthMethodOIDC, want: idleOIDC},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeRepository{}
			userID := uuid.New()

			if _, err := frozenSvc(t, repo, now).Create(context.Background(), sessions.CreateSessionOpts{
				UserID:     userID,
				AuthMethod: tc.method,
			}); err != nil {
				t.Fatalf("Create: %v", err)
			}

			if got := repo.gotInsert.ExpiresAt; !got.Equal(now.Add(tc.want)) {
				t.Errorf("ExpiresAt = %v, want %v", got, now.Add(tc.want))
			}

			if repo.gotInsert.UserID != userID {
				t.Errorf("UserID = %v, want %v", repo.gotInsert.UserID, userID)
			}

			if repo.gotInsert.AuthMethod != tc.method {
				t.Errorf("AuthMethod = %v, want %v", repo.gotInsert.AuthMethod, tc.method)
			}
		})
	}
}

func TestCreateRejectsUnknownAuthMethod(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}

	got, err := frozenSvc(t, repo, time.Now()).Create(context.Background(), sessions.CreateSessionOpts{
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethod("sms"),
	})
	if err == nil {
		t.Fatal("Create doit refuser une méthode inconnue")
	}

	if got != nil {
		t.Errorf("Create a renvoyé %+v en plus de l'erreur", got)
	}

	if repo.inserts != 0 {
		t.Errorf("Insert appelée %d fois pour une méthode inconnue", repo.inserts)
	}
}

func TestCreatePropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("disk full")
	repo := &fakeRepository{insertErr: sentinel}

	got, err := frozenSvc(t, repo, time.Now()).Create(context.Background(), sessions.CreateSessionOpts{
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
	})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}

	if got != nil {
		t.Errorf("Create a renvoyé %+v en plus de l'erreur", got)
	}
}

func TestNewDefaultsNowToTimeNow(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}

	svc, err := sessions.NewService(validConfig(), sessions.ServiceDeps{Repository: repo})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	before := time.Now()

	if _, err := svc.Create(context.Background(), sessions.CreateSessionOpts{
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
	}); err != nil {
		t.Fatalf("Create: %v", err)
	}

	if repo.gotInsert.ExpiresAt.Before(before.Add(idlePwd)) {
		t.Errorf("ExpiresAt = %v, l'horloge par défaut n'est pas time.Now", repo.gotInsert.ExpiresAt)
	}
}

func aliveSession(now time.Time, left time.Duration, method sessions.AuthMethod) *sessions.Session {
	return &sessions.Session{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		AuthMethod: method,
		CreatedAt:  now.Add(-time.Hour),
		ExpiresAt:  now.Add(left),
	}
}

func TestAuthenticateReturnsUserAndSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	session := aliveSession(now, idlePwd, sessions.AuthMethodPassword)
	user := &users.User{ID: session.UserID, Name: "alice"}
	repo := &fakeRepository{session: session, user: user}

	got, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if got.User != user {
		t.Errorf("User = %+v, want %+v", got.User, user)
	}

	if got.Session.ID != session.ID {
		t.Errorf("Session.ID = %v, want %v", got.Session.ID, session.ID)
	}

	want := sha256.Sum256([]byte("letoken"))
	if !bytes.Equal(repo.gotHash, want[:]) {
		t.Errorf("GetByTokenHash a reçu %x, want %x", repo.gotHash, want)
	}
}

func TestAuthenticateRejectsExpiredSession(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)

	tests := map[string]time.Duration{
		"expirée depuis une seconde": -time.Second,
		"expirée à la seconde près":  0,
	}

	for name, left := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			repo := &fakeRepository{
				session: aliveSession(now, left, sessions.AuthMethodPassword),
				user:    &users.User{ID: uuid.New()},
			}

			got, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken")
			if !errors.Is(err, sessions.ErrInvalidSession) {
				t.Errorf("err = %v, want ErrInvalidSession", err)
			}

			if got != nil {
				t.Errorf("Authenticate a renvoyé %+v en plus de l'erreur", got)
			}
		})
	}
}

func TestAuthenticateRejectsUnknownAuthMethodFromDatabase(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	repo := &fakeRepository{
		session: aliveSession(now, idlePwd, sessions.AuthMethod("webauthn")),
		user:    &users.User{ID: uuid.New()},
	}

	got, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken")
	if !errors.Is(err, sessions.ErrInvalidSession) {
		t.Errorf("err = %v, want ErrInvalidSession", err)
	}

	if got != nil {
		t.Errorf("Authenticate a renvoyé %+v en plus de l'erreur", got)
	}

	if repo.updates != 0 {
		t.Errorf("UpdateExpiry appelée %d fois pour une session rejetée", repo.updates)
	}
}

func TestAuthenticateRejectsUnknownToken(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{getErr: domain.ErrNotFound}

	got, err := frozenSvc(t, repo, time.Now()).Authenticate(context.Background(), "letoken")
	if !errors.Is(err, sessions.ErrInvalidSession) {
		t.Errorf("err = %v, want ErrInvalidSession", err)
	}

	if got != nil {
		t.Errorf("Authenticate a renvoyé %+v en plus de l'erreur", got)
	}
}

func TestAuthenticateRejectsEmptyTokenWithoutTouchingRepository(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}

	if _, err := frozenSvc(t, repo, time.Now()).Authenticate(context.Background(), ""); !errors.Is(
		err, sessions.ErrInvalidSession,
	) {
		t.Errorf("err = %v, want ErrInvalidSession", err)
	}

	if repo.gets != 0 {
		t.Errorf("GetByTokenHash appelée %d fois pour un token vide", repo.gets)
	}
}

func TestAuthenticateDoesNotMaskInfrastructureErrors(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection refused")
	repo := &fakeRepository{getErr: sentinel}

	_, err := frozenSvc(t, repo, time.Now()).Authenticate(context.Background(), "letoken")
	if errors.Is(err, sessions.ErrInvalidSession) {
		t.Error("une panne SQL a été traduite en ErrInvalidSession")
	}

	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}
}

func TestAuthenticateSkipsRenewalBelowThreshold(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	repo := &fakeRepository{
		session: aliveSession(now, idlePwd-renew/2, sessions.AuthMethodPassword),
		user:    &users.User{ID: uuid.New()},
	}

	got, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if repo.updates != 0 {
		t.Errorf("UpdateExpiry appelée %d fois en deçà du seuil", repo.updates)
	}

	if !got.Session.ExpiresAt.Equal(now.Add(idlePwd - renew/2)) {
		t.Errorf("ExpiresAt = %v, want inchangée", got.Session.ExpiresAt)
	}
}

func TestAuthenticateRenewsBeyondThreshold(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	session := aliveSession(now, time.Hour, sessions.AuthMethodPassword)
	repo := &fakeRepository{session: session, user: &users.User{ID: uuid.New()}}

	got, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if repo.updates != 1 {
		t.Fatalf("UpdateExpiry appelée %d fois, want 1", repo.updates)
	}

	want := now.Add(idlePwd)
	if !repo.gotExpiry.Equal(want) {
		t.Errorf("UpdateExpiry a reçu %v, want %v", repo.gotExpiry, want)
	}

	if repo.gotID != session.ID {
		t.Errorf("UpdateExpiry a reçu l'ID %v, want %v", repo.gotID, session.ID)
	}

	if !got.Session.ExpiresAt.Equal(want) {
		t.Errorf("ExpiresAt renvoyée = %v, want %v", got.Session.ExpiresAt, want)
	}
}

func TestAuthenticateClampsRenewalToAbsoluteLimit(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	session := &sessions.Session{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
		CreatedAt:  now.Add(-89 * day),
		ExpiresAt:  now.Add(time.Hour),
	}
	repo := &fakeRepository{session: session, user: &users.User{ID: session.UserID}}

	if _, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	want := session.CreatedAt.Add(absPwd)
	if !repo.gotExpiry.Equal(want) {
		t.Errorf("UpdateExpiry a reçu %v, want le plafond %v", repo.gotExpiry, want)
	}
}

func TestAuthenticateSkipsRenewalWhenClampDoesNotAdvance(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	created := now.Add(-absPwd).Add(time.Hour)
	session := &sessions.Session{
		ID:         uuid.New(),
		UserID:     uuid.New(),
		AuthMethod: sessions.AuthMethodPassword,
		CreatedAt:  created,
		ExpiresAt:  created.Add(absPwd),
	}
	repo := &fakeRepository{session: session, user: &users.User{ID: session.UserID}}

	if _, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if repo.updates != 0 {
		t.Errorf("UpdateExpiry appelée %d fois alors que le plafond ne fait rien avancer", repo.updates)
	}
}

func TestAuthenticateSucceedsWhenRenewalFails(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	sentinel := errors.New("deadlock detected")
	session := aliveSession(now, time.Hour, sessions.AuthMethodPassword)
	repo := &fakeRepository{
		session:   session,
		user:      &users.User{ID: session.UserID},
		updateErr: sentinel,
	}

	got, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if !errors.Is(got.RenewErr, sentinel) {
		t.Errorf("RenewErr = %v, want %v", got.RenewErr, sentinel)
	}

	if !got.Session.ExpiresAt.Equal(session.ExpiresAt) {
		t.Errorf("ExpiresAt = %v, want l'ancienne valeur %v", got.Session.ExpiresAt, session.ExpiresAt)
	}
}

func TestAuthenticateRenewsWithTTLOfAuthMethod(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 3, 12, 0, 0, 0, time.UTC)
	session := aliveSession(now, 30*time.Minute, sessions.AuthMethodOIDC)
	repo := &fakeRepository{session: session, user: &users.User{ID: session.UserID}}

	if _, err := frozenSvc(t, repo, now).Authenticate(context.Background(), "letoken"); err != nil {
		t.Fatalf("Authenticate: %v", err)
	}

	if want := now.Add(idleOIDC); !repo.gotExpiry.Equal(want) {
		t.Errorf("UpdateExpiry a reçu %v, want %v — le TTL oidc n'a pas été appliqué", repo.gotExpiry, want)
	}
}

func TestRevokeDeletesByHash(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}

	if err := frozenSvc(t, repo, time.Now()).Revoke(context.Background(), "letoken"); err != nil {
		t.Fatalf("Revoke: %v", err)
	}

	want := sha256.Sum256([]byte("letoken"))
	if !bytes.Equal(repo.gotHash, want[:]) {
		t.Errorf("DeleteByTokenHash a reçu %x, want %x", repo.gotHash, want)
	}
}

func TestRevokeRejectsEmptyTokenWithoutTouchingRepository(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}

	if err := frozenSvc(t, repo, time.Now()).Revoke(context.Background(), ""); !errors.Is(
		err, sessions.ErrInvalidSession,
	) {
		t.Errorf("err = %v, want ErrInvalidSession", err)
	}

	if repo.deleteHashes != 0 {
		t.Errorf("DeleteByTokenHash appelée %d fois pour un token vide", repo.deleteHashes)
	}
}

func TestRevokePropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection reset")
	repo := &fakeRepository{deleteErr: sentinel}

	err := frozenSvc(t, repo, time.Now()).Revoke(context.Background(), "letoken")
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}
}

func TestRevokeAllForUserDeletesByUserID(t *testing.T) {
	t.Parallel()

	repo := &fakeRepository{}
	userID := uuid.New()

	if err := frozenSvc(t, repo, time.Now()).RevokeAllForUser(context.Background(), userID); err != nil {
		t.Fatalf("RevokeAllForUser: %v", err)
	}

	if repo.gotUserID != userID {
		t.Errorf("DeleteByUserID a reçu %v, want %v", repo.gotUserID, userID)
	}
}

func TestRevokeAllForUserPropagatesRepositoryError(t *testing.T) {
	t.Parallel()

	sentinel := errors.New("connection reset")
	repo := &fakeRepository{deleteErr: sentinel}

	err := frozenSvc(t, repo, time.Now()).RevokeAllForUser(context.Background(), uuid.New())
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v, l'erreur d'origine n'est plus atteignable", err)
	}
}

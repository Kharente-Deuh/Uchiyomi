// SPDX-License-Identifier: AGPL-3.0-or-later

package pg_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders/repository/pg"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction/pgtx"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	colDisplayName     = "display_name"
	colScopes          = "scopes"
	colClientSecretEnc = "client_secret_enc"
	testDisplayName    = "Keycloak"
	scopeOpenID        = "openid"
	scopeProfile       = "profile"
	adminGroupValue    = "admins"
	testIssuerURL      = "https://sso.example.com"
)

func duplicateKeyErr() error {
	return &pgconn.PgError{
		Code:    "23505",
		Message: `duplicate key value violates unique constraint "idx_oidc_providers_issuer_url"`,
	}
}

func newRepo(t *testing.T) (*pg.PGOIDCProvidersRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	return r, mock
}

func providerRows(id uuid.UUID) *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", colDisplayName, "issuer_url", "client_id", colClientSecretEnc,
		colScopes, "username_claim", "role_claim", "admin_values",
		"allowed_values", "auto_provision",
	}).AddRow(
		id, testDisplayName, testIssuerURL, "uchiyomi", []byte("enc"),
		pq.StringArray{scopeOpenID, scopeProfile}, "preferred_username",
		"groups", pq.StringArray{adminGroupValue},
		pq.StringArray{"users"},
		true,
	)
}

func TestCreateJoinsAmbientTransaction(t *testing.T) {
	t.Parallel()

	db, mock := pgtest.New(t)

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tr, err := pgtx.New(pgtx.Deps{DB: db})
	if err != nil {
		t.Fatalf("pgtx.New: %v", err)
	}

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "oidc_providers"`).
		WillReturnRows(sqlmock.NewRows([]string{"scopes", "id"}).AddRow(nil, uuid.New()))
	mock.ExpectCommit()

	err = tr.WithinTx(context.Background(), transaction.TxOpts{}, func(ctx context.Context) error {
		_, createErr := r.Create(ctx, oidcproviders.CreateOIDCProviderOpts{DisplayName: "keycloak", IssuerURL: "https://idp"})

		if createErr != nil {
			return fmt.Errorf("r.Create: %w", createErr)
		}

		return nil
	})
	if err != nil {
		t.Fatalf("WithinTx: %v", err)
	}
}

func TestNewValidatesDeps(t *testing.T) {
	t.Parallel()

	if _, err := pg.New(pg.Deps{}); err == nil || err.Error() != "deps.Validate: db is required" {
		t.Errorf("pg.New(pg.Deps{}) = %v, want %q", err, "deps.Validate: db is required")
	}
}

func TestGetByID(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "oidc_providers" WHERE id = \$1`).
		WithArgs(id, 1).
		WillReturnRows(providerRows(id))

	got, err := r.GetByID(context.Background(), id)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}

	if got.ID != id || got.DisplayName != testDisplayName || got.IssuerURL != testIssuerURL {
		t.Errorf("GetByID() = %+v", got)
	}

	if !reflect.DeepEqual(got.Scopes, []string{scopeOpenID, scopeProfile}) {
		t.Errorf("Scopes = %v, want [openid profile]", got.Scopes)
	}

	if !reflect.DeepEqual(got.AdminValues, []string{adminGroupValue}) {
		t.Errorf("AdminValues = %v, want [admins]", got.AdminValues)
	}

	if got.RoleClaim == nil || *got.RoleClaim != "groups" {
		t.Errorf("RoleClaim = %v, want \"groups\"", got.RoleClaim)
	}

	if !got.AutoProvision {
		t.Error("AutoProvision = false, want true")
	}
}

func TestGetByIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "oidc_providers"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	if _, err := r.GetByID(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByID = %v, want domain.ErrNotFound", err)
	}
}

func TestGetByIDError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection refused")
	mock.ExpectQuery(`SELECT \* FROM "oidc_providers"`).WillReturnError(sentinel)

	_, err := r.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}

	if errors.Is(err, domain.ErrNotFound) {
		t.Error("une panne SQL ne doit pas être traduite en ErrNotFound")
	}
}

func TestGetByIssuerURL(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()

	mock.ExpectQuery(`SELECT \* FROM "oidc_providers" WHERE issuer_url = \$1`).
		WithArgs(testIssuerURL, 1).
		WillReturnRows(providerRows(id))

	got, err := r.GetByIssuerURL(context.Background(), testIssuerURL)
	if err != nil {
		t.Fatalf("GetByIssuerURL: %v", err)
	}

	if got.ID != id {
		t.Errorf("GetByIssuerURL() = %+v", got)
	}
}

func TestGetByIssuerURLNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`SELECT \* FROM "oidc_providers"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}))

	_, err := r.GetByIssuerURL(context.Background(), "https://inconnu")
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("GetByIssuerURL = %v, want domain.ErrNotFound", err)
	}
}

func TestGetAll(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id1, id2 := uuid.New(), uuid.New()
	testCreatedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	mock.ExpectQuery(`COUNT\(DISTINCT federated_identities.user_id\).*LEFT JOIN federated_identities`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", colDisplayName, "created_at", "user_count"}).
				AddRow(id1, testDisplayName, testCreatedAt, 2).
				AddRow(id2, "Authentik", testCreatedAt, 0),
		)

	got, err := r.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	want := []oidcproviders.LightOIDCProvider{
		{ID: id1, DisplayName: testDisplayName, CreatedAt: testCreatedAt, UserCount: 2},
		{ID: id2, DisplayName: "Authentik", CreatedAt: testCreatedAt, UserCount: 0},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetAll() = %+v, want %+v", got, want)
	}
}

func TestGetAllEmpty(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`LEFT JOIN federated_identities`).
		WillReturnRows(sqlmock.NewRows([]string{"id", colDisplayName}))

	got, err := r.GetAll(context.Background())
	if err != nil {
		t.Fatalf("GetAll: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("GetAll() = %+v, want vide", got)
	}
}

func TestGetAllError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("timeout")
	mock.ExpectQuery(`LEFT JOIN federated_identities`).WillReturnError(sentinel)

	if _, err := r.GetAll(context.Background()); !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}
}

func TestGetUsers(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id1, id2 := uuid.New(), uuid.New()
	linkedAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)

	mock.ExpectQuery(`JOIN users ON users.id = federated_identities.user_id`).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "username", "is_admin", "linked_at"}).
				AddRow(id1, "alice", true, linkedAt).
				AddRow(id2, "bob", false, linkedAt),
		)

	got, err := r.GetUsers(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}

	want := []oidcproviders.OIDCProviderUser{
		{ID: id1, Username: "alice", LinkedAt: linkedAt, IsAdmin: true},
		{ID: id2, Username: "bob", LinkedAt: linkedAt, IsAdmin: false},
	}

	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetUsers() = %+v, want %+v", got, want)
	}
}

func TestGetUsersEmpty(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectQuery(`JOIN users ON users.id = federated_identities.user_id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "username", "is_admin", "linked_at"}))

	got, err := r.GetUsers(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("GetUsers: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("GetUsers() = %+v, want vide", got)
	}
}

func TestGetUsersError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("timeout")
	mock.ExpectQuery(`JOIN users ON users.id = federated_identities.user_id`).WillReturnError(sentinel)

	if _, err := r.GetUsers(context.Background(), uuid.New()); !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}
}

func TestCreate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "oidc_providers"`).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(uuid.New()))
	mock.ExpectCommit()

	roleClaim := "groups"

	got, err := r.Create(context.Background(), oidcproviders.CreateOIDCProviderOpts{
		DisplayName:     testDisplayName,
		IssuerURL:       testIssuerURL,
		ClientID:        "uchiyomi",
		ClientSecretEnc: []byte("enc"),
		Scopes:          []string{scopeOpenID, scopeProfile, "email"},
		UsernameClaim:   "preferred_username",
		RoleClaim:       &roleClaim,
		AdminValues:     []string{adminGroupValue},
		AutoProvision:   true,
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	if got.ID == uuid.Nil {
		t.Error("Create n'a pas généré d'ID")
	}

	if got.DisplayName != testDisplayName || got.IssuerURL != testIssuerURL {
		t.Errorf("Create() = %+v", got)
	}

	if !reflect.DeepEqual(got.Scopes, []string{scopeOpenID, scopeProfile, "email"}) {
		t.Errorf("Scopes = %v", got.Scopes)
	}

	if got.CreatedAt.IsZero() || got.UpdatedAt.IsZero() {
		t.Errorf("horodatages non renseignés: created=%v updated=%v", got.CreatedAt, got.UpdatedAt)
	}
}

func TestCreateDuplicate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "oidc_providers"`).WillReturnError(duplicateKeyErr())
	mock.ExpectRollback()

	got, err := r.Create(context.Background(), oidcproviders.CreateOIDCProviderOpts{IssuerURL: testIssuerURL})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Create = %v, want domain.ErrAlreadyExists", err)
	}

	if got != nil {
		t.Errorf("Create a renvoyé %+v en plus de l'erreur", got)
	}
}

func TestCreateError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("disk full")
	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO "oidc_providers"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	_, err := r.Create(context.Background(), oidcproviders.CreateOIDCProviderOpts{})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}

	if errors.Is(err, domain.ErrAlreadyExists) {
		t.Error("une panne SQL ne doit pas être traduite en ErrAlreadyExists")
	}
}

func TestUpdateWritesTheSecretWhenGiven(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "oidc_providers" SET .*"client_secret_enc"`).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "oidc_providers" WHERE id = \$1`).
		WithArgs(id, 1).
		WillReturnRows(providerRows(id))

	got, err := r.Update(context.Background(), id, oidcproviders.UpdateOIDCProviderOpts{
		DisplayName:     testDisplayName,
		IssuerURL:       testIssuerURL,
		ClientID:        "uchiyomi",
		ClientSecretEnc: []byte("newenc"),
		Scopes:          []string{scopeOpenID},
		UsernameClaim:   "preferred_username",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	if got.ID != id {
		t.Errorf("Update() = %+v", got)
	}
}

func TestUpdateLeavesTheSecretAloneWhenOmitted(t *testing.T) {
	t.Parallel()

	var updateSQL string

	sqlDB, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherFunc(
		func(expectedSQL, actualSQL string) error {
			if strings.HasPrefix(actualSQL, `UPDATE "oidc_providers"`) {
				updateSQL = actualSQL
			}

			matched, matchErr := regexp.MatchString(expectedSQL, actualSQL)
			if matchErr != nil {
				return fmt.Errorf("regexp.MatchString: %w", matchErr)
			}

			if !matched {
				return fmt.Errorf("actual sql %q does not match expected regexp %q", actualSQL, expectedSQL)
			}

			return nil
		},
	)))
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("sqlmock expectations not met: %v", err)
		}

		sqlDB.Close()
	})

	db, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}),
		&gorm.Config{TranslateError: true},
	)
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	r, err := pg.New(pg.Deps{DB: db})
	if err != nil {
		t.Fatalf("pg.New: %v", err)
	}

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "oidc_providers" SET`).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	mock.ExpectQuery(`SELECT \* FROM "oidc_providers"`).WillReturnRows(providerRows(id))

	if _, err := r.Update(context.Background(), id, oidcproviders.UpdateOIDCProviderOpts{
		DisplayName: testDisplayName,
		IssuerURL:   testIssuerURL,
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	if strings.Contains(updateSQL, colClientSecretEnc) {
		t.Errorf("Update() SET clause included %s: %s", colClientSecretEnc, updateSQL)
	}
}

func TestUpdateNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "oidc_providers"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	_, err := r.Update(context.Background(), uuid.New(), oidcproviders.UpdateOIDCProviderOpts{})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Update = %v, want domain.ErrNotFound", err)
	}
}

func TestUpdateDuplicate(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "oidc_providers"`).WillReturnError(duplicateKeyErr())
	mock.ExpectRollback()

	_, err := r.Update(context.Background(), uuid.New(), oidcproviders.UpdateOIDCProviderOpts{IssuerURL: testIssuerURL})
	if !errors.Is(err, domain.ErrAlreadyExists) {
		t.Errorf("Update = %v, want domain.ErrAlreadyExists", err)
	}
}

func TestUpdateError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("disk full")
	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE "oidc_providers"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	_, err := r.Update(context.Background(), uuid.New(), oidcproviders.UpdateOIDCProviderOpts{})
	if !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}

	if errors.Is(err, domain.ErrAlreadyExists) {
		t.Error("a plain SQL failure must not be translated into domain.ErrAlreadyExists")
	}

	if errors.Is(err, domain.ErrNotFound) {
		t.Error("a plain SQL failure must not be translated into domain.ErrNotFound")
	}
}

func TestDeleteByID(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	id := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "oidc_providers" WHERE id = \$1`).
		WithArgs(id).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	if err := r.DeleteByID(context.Background(), id); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}

func TestDeleteByIDNotFound(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "oidc_providers"`).WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	if err := r.DeleteByID(context.Background(), uuid.New()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("Delete = %v, want domain.ErrNotFound", err)
	}
}

func TestDeleteByIDError(t *testing.T) {
	t.Parallel()

	r, mock := newRepo(t)

	sentinel := errors.New("connection reset")
	mock.ExpectBegin()
	mock.ExpectExec(`DELETE FROM "oidc_providers"`).WillReturnError(sentinel)
	mock.ExpectRollback()

	if err := r.DeleteByID(context.Background(), uuid.New()); !errors.Is(err, sentinel) {
		t.Errorf("err = %v", err)
	}
}

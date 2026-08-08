// SPDX-License-Identifier: AGPL-3.0-or-later

package database

import (
	"context"
	"errors"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgtest"
)

func TestPingSucceedsWhenDatabaseAnswers(t *testing.T) {
	db, mock := pgtest.NewWithPings(t)
	mock.ExpectPing()

	pgdb := &PGDB{DB: db}

	if err := pgdb.Ping(context.Background()); err != nil {
		t.Fatalf("Ping: %v", err)
	}
}

func TestPingFailsWhenDatabaseDoesNot(t *testing.T) {
	wantErr := errors.New("connexion refusée")

	db, mock := pgtest.NewWithPings(t)
	mock.ExpectPing().WillReturnError(wantErr)

	pgdb := &PGDB{DB: db}

	err := pgdb.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping: erreur attendue, nil obtenu")
	}

	if !errors.Is(err, wantErr) {
		t.Errorf("Ping: erreur %v, attendu qu'elle encapsule %v", err, wantErr)
	}
}

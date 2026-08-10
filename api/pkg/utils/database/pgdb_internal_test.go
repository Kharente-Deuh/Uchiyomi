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
	wantErr := errors.New("connection refused")

	db, mock := pgtest.NewWithPings(t)
	mock.ExpectPing().WillReturnError(wantErr)

	pgdb := &PGDB{DB: db}

	err := pgdb.Ping(context.Background())
	if err == nil {
		t.Fatal("Ping: error expected, got nil")
	}

	if !errors.Is(err, wantErr) {
		t.Errorf("Ping: error %v, want it to wrap %v", err, wantErr)
	}
}

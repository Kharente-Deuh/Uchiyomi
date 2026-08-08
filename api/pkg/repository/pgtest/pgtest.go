// SPDX-License-Identifier: AGPL-3.0-or-later

package pgtest

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func New(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	return newDB(t, false)
}

func NewWithPings(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	return newDB(t, true)
}

func newDB(t *testing.T, monitorPings bool) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(
		sqlmock.MonitorPingsOption(monitorPings),
		sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp),
	)
	if err != nil {
		t.Fatalf("sqlmock.New: %v", err)
	}

	t.Cleanup(func() {
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("attentes sqlmock non satisfaites: %v", err)
		}

		sqlDB.Close()
	})

	db, err := gorm.Open(
		postgres.New(postgres.Config{Conn: sqlDB, PreferSimpleProtocol: true}),
		&gorm.Config{
			TranslateError:       true,
			Logger:               logger.Default.LogMode(logger.Silent),
			DisableAutomaticPing: true,
		},
	)
	if err != nil {
		t.Fatalf("gorm.Open: %v", err)
	}

	return db, mock
}

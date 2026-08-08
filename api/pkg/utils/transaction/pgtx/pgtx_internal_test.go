// SPDX-License-Identifier: AGPL-3.0-or-later

package pgtx

import (
	"database/sql"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/transaction"
)

func TestTxOptions(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		opts transaction.TxOpts
		want sql.IsolationLevel
	}{
		"par défaut, on laisse le niveau du serveur": {
			opts: transaction.TxOpts{},
			want: sql.LevelDefault,
		},
		"serializable est traduit explicitement": {
			opts: transaction.TxOpts{Isolation: transaction.IsolationSerializable},
			want: sql.LevelSerializable,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := txOptions(tc.opts)
			if got.Isolation != tc.want {
				t.Errorf("txOptions(%+v).Isolation = %v, want %v", tc.opts, got.Isolation, tc.want)
			}
		})
	}
}

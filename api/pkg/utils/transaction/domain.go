// SPDX-License-Identifier: AGPL-3.0-or-later

package transaction

import "context"

type IsolationLevel int

const (
	IsolationDefault IsolationLevel = iota
	IsolationSerializable
)

type TxOpts struct {
	Isolation IsolationLevel
}

type Transactor interface {
	WithinTx(ctx context.Context, opts TxOpts, fn func(ctx context.Context) error) error
}

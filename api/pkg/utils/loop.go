// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"context"
	"time"
)

type LoopOpts struct {
	Fn       func(context.Context) error
	Interval time.Duration
}

func Loop(ctx context.Context, opts LoopOpts) error {
	timer := time.NewTimer(opts.Interval)
	defer timer.Stop()

	for {
		if err := ctx.Err(); err != nil {
			//nolint:wrapcheck
			return err
		}

		if err := opts.Fn(ctx); err != nil {
			//nolint:wrapcheck
			return err
		}

		timer.Reset(opts.Interval)

		select {
		case <-ctx.Done():
			//nolint:wrapcheck
			return ctx.Err()
		case <-timer.C:
		}
	}
}

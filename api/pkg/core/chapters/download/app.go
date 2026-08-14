// SPDX-License-Identifier: AGPL-3.0-or-later

package download

import "context"

type App struct {
	worker *Worker
}

func NewApp(worker *Worker) (*App, error) {
	if worker == nil {
		return nil, ErrWorkerRequired
	}

	return &App{worker: worker}, nil
}

func (a *App) Run(ctx context.Context) error {
	//nolint:wrapcheck
	return a.worker.Run(ctx)
}

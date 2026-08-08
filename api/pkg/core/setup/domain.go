// SPDX-License-Identifier: AGPL-3.0-or-later

package setup

import (
	"context"
	"errors"

	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/sessions"
)

type SetupService interface {
	IsSetupRequired(context.Context) (bool, error)
	DoSetup(context.Context, DoSetupOpts) (*sessions.IssuedSession, error)
}

type DoSetupOpts struct {
	Username string
	Password string
}

var (
	ErrSetupNotNeeded   = errors.New("setup not needed")
	ErrSessionNotIssued = errors.New("session not issued")
)

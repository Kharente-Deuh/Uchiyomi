// SPDX-License-Identifier: AGPL-3.0-or-later

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/auth/oidcproviders"
	"github.com/kharente-deuh/uchiyomi-server/pkg/core/domain"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/logging"
)

type providersService interface {
	List(ctx context.Context) ([]oidcproviders.LightOIDCProvider, error)
	GetByID(ctx context.Context, id uuid.UUID) (*oidcproviders.OIDCProvider, error)
	Create(ctx context.Context, opts oidcproviders.CreateOpts) (*oidcproviders.OIDCProvider, error)
	Update(ctx context.Context, id uuid.UUID, opts oidcproviders.UpdateOpts) (*oidcproviders.OIDCProvider, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Probe(ctx context.Context, issuerURL string) (*oidcproviders.ProbeResult, error)
}

type Config struct {
	Endpoint    string
	Middlewares chi.Middlewares
}

//nolint:dupl
func (cfg *Config) Validate() error {
	if cfg.Endpoint == "" {
		return errors.New("endpoint is required")
	}

	if !strings.HasPrefix(cfg.Endpoint, "/") {
		return fmt.Errorf("endpoint must start with '/', got %q", cfg.Endpoint)
	}

	hasNilMiddlewares := slices.ContainsFunc(cfg.Middlewares, func(m func(http.Handler) http.Handler) bool {
		return m == nil
	})

	if hasNilMiddlewares {
		return errors.New("all middlewares must not be nil")
	}

	return nil
}

type Deps struct {
	Service providersService
	Logger  *slog.Logger
}

func (deps *Deps) Validate() error {
	if deps.Service == nil {
		return errors.New("service is required")
	}

	if deps.Logger == nil {
		return errors.New("logger is required")
	}

	return nil
}

type Controller struct {
	deps Deps
	cfg  Config
}

func New(cfg Config, deps Deps) (*Controller, error) {
	var err error
	if err = cfg.Validate(); err != nil {
		return nil, fmt.Errorf("cfg.Validate: %w", err)
	}

	if err = deps.Validate(); err != nil {
		return nil, fmt.Errorf("deps.Validate: %w", err)
	}

	deps.Logger = deps.Logger.With("component", "oidcproviders.gateway.http")

	return &Controller{cfg: cfg, deps: deps}, nil
}

func (c *Controller) InitRouter(r chi.Router) {
	r.Route(c.cfg.Endpoint, func(r chi.Router) {
		for _, m := range c.cfg.Middlewares {
			r.Use(m)
		}

		r.Get("/", c.list)
		r.Post("/", c.create)
		r.Post("/probe", c.probe)
		r.Get("/{id}", c.get)
		r.Put("/{id}", c.update)
		r.Delete("/{id}", c.delete)
	})
}

func (c *Controller) list(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	providers, err := c.deps.Service.List(ctx)
	if err != nil {
		c.writeServiceError(w, r, "failed to list oidc providers", err)

		return
	}

	res := make([]LightProviderResponse, 0, len(providers))
	for _, p := range providers {
		res = append(res, LightProviderResponse{ID: p.ID.String(), DisplayName: p.DisplayName, IssuerURL: p.IssuerURL})
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, res)
}

func (c *Controller) get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	provider, err := c.deps.Service.GetByID(ctx, id)
	if err != nil {
		c.writeServiceError(w, r, "failed to read an oidc provider", err)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, toProviderResponse(provider))
}

func (c *Controller) create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httputils.DecodeJSON[CreateProviderRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, err.Error())

		return
	}

	provider, err := c.deps.Service.Create(ctx, oidcproviders.CreateOpts{
		DisplayName:   req.DisplayName.String(),
		IssuerURL:     req.IssuerURL.String(),
		ClientID:      req.ClientID.String(),
		ClientSecret:  req.ClientSecret.String(),
		UsernameClaim: req.UsernameClaim.String(),
		Scopes:        req.Scopes,
		RoleClaim:     req.RoleClaim,
		AdminValues:   req.AdminValues,
		AllowedValues: req.AllowedValues,
		AutoProvision: req.AutoProvision,
	})
	if err != nil {
		c.writeServiceError(w, r, "failed to create an oidc provider", err)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusCreated, toProviderResponse(provider))
}

func (c *Controller) update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	req, err := httputils.DecodeJSON[UpdateProviderRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, err.Error())

		return
	}

	var secret *string

	if req.ClientSecret != nil {
		value := req.ClientSecret.String()
		secret = &value
	}

	provider, err := c.deps.Service.Update(ctx, id, oidcproviders.UpdateOpts{
		ClientSecret:  secret,
		DisplayName:   req.DisplayName.String(),
		IssuerURL:     req.IssuerURL.String(),
		ClientID:      req.ClientID.String(),
		UsernameClaim: req.UsernameClaim.String(),
		Scopes:        req.Scopes,
		RoleClaim:     req.RoleClaim,
		AdminValues:   req.AdminValues,
		AllowedValues: req.AllowedValues,
		AutoProvision: req.AutoProvision,
	})
	if err != nil {
		c.writeServiceError(w, r, "failed to update an oidc provider", err)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, toProviderResponse(provider))
}

func (c *Controller) delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "id must be a valid UUID")

		return
	}

	if err := c.deps.Service.Delete(ctx, id); err != nil {
		c.writeServiceError(w, r, "failed to delete an oidc provider", err)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (c *Controller) probe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req, err := httputils.DecodeJSON[ProbeRequest](r)
	if err != nil {
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, err.Error())

		return
	}

	result, err := c.deps.Service.Probe(ctx, req.IssuerURL.String())
	if err != nil {
		c.writeServiceError(w, r, "failed to probe an oidc issuer", err)

		return
	}

	httputils.WriteJSON(w, c.deps.Logger, http.StatusOK, ProbeResponse{
		Issuer:                    result.Issuer,
		AuthorizationEndpoint:     result.AuthorizationEndpoint,
		TokenEndpoint:             result.TokenEndpoint,
		UserInfoEndpoint:          result.UserInfoEndpoint,
		EndSessionEndpoint:        result.EndSessionEndpoint,
		RedirectURI:               result.RedirectURI,
		SupportsRPInitiatedLogout: result.SupportsRPInitiatedLogout,
	})
}

func (c *Controller) writeServiceError(w http.ResponseWriter, r *http.Request, msg string, err error) {
	switch {
	case errors.Is(err, oidcproviders.ErrUnreachableIssuer):
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest, "issuer is unreachable")
	case errors.Is(err, oidcproviders.ErrIncompleteIssuer):
		httputils.WriteError(w, c.deps.Logger, http.StatusBadRequest,
			"issuer discovery document does not advertise the required endpoints")
	case errors.Is(err, domain.ErrAlreadyExists):
		httputils.WriteError(w, c.deps.Logger, http.StatusConflict, "issuer URL is already declared")
	case errors.Is(err, domain.ErrNotFound):
		httputils.WriteError(w, c.deps.Logger, http.StatusNotFound, "")
	default:
		c.deps.Logger.ErrorContext(r.Context(), msg, logging.Err(err))
		httputils.WriteError(w, c.deps.Logger, http.StatusInternalServerError, "")
	}
}

func toProviderResponse(p *oidcproviders.OIDCProvider) ProviderResponse {
	return ProviderResponse{
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
		RoleClaim:     p.RoleClaim,
		ID:            p.ID.String(),
		DisplayName:   p.DisplayName,
		IssuerURL:     p.IssuerURL,
		ClientID:      p.ClientID,
		UsernameClaim: p.UsernameClaim,
		Scopes:        p.Scopes,
		AdminValues:   p.AdminValues,
		AllowedValues: p.AllowedValues,
		AutoProvision: p.AutoProvision,
	}
}

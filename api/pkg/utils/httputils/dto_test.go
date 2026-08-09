// SPDX-License-Identifier: AGPL-3.0-or-later

package httputils_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils/httputils"
)

type taggedRequest struct {
	IssuerURL     string `json:"issuerUrl"     validate:"required,url"`
	ClientSecret  string `json:"clientSecret"  validate:"required,min=8"`
	UsernameClaim string `json:"usernameClaim,omitempty" validate:"omitempty,oneof=sub email"`
}

type untaggedRequest struct {
	IssuerURL string `validate:"required,url"`
}

type skippedRequest struct {
	Internal string `json:"-" validate:"required"`
}

func decode[T any](t *testing.T, body string) (*T, error) {
	t.Helper()

	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	return httputils.DecodeJSON[T](req)
}

func TestDecodeJSONNamesTheJSONKey(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		body    string
		wantErr string
	}{
		"url invalide": {
			body:    `{"issuerUrl":"nope","clientSecret":"averylongsecret"}`,
			wantErr: "issuerUrl must be a valid URL",
		},
		"champ manquant": {
			body:    `{"issuerUrl":"https://idp.example.com"}`,
			wantErr: "clientSecret is required",
		},
		"longueur minimale": {
			body:    `{"issuerUrl":"https://idp.example.com","clientSecret":"short"}`,
			wantErr: "clientSecret must be at least 8 characters",
		},
		"valeur hors liste": {
			body:    `{"issuerUrl":"https://idp.example.com","clientSecret":"averylongsecret","usernameClaim":"name"}`,
			wantErr: "usernameClaim must be one of: sub, email",
		},
		"erreurs cumulées": {
			body:    `{}`,
			wantErr: "issuerUrl is required, clientSecret is required",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := decode[taggedRequest](t, tc.body)
			if err == nil {
				t.Fatalf("DecodeJSON(%s) = nil, want %q", tc.body, tc.wantErr)
			}

			if got := err.Error(); got != tc.wantErr {
				t.Errorf("err = %q, want %q", got, tc.wantErr)
			}
		})
	}
}

func TestDecodeJSONFallsBackToTheGoFieldWithoutJSONTag(t *testing.T) {
	t.Parallel()

	_, err := decode[untaggedRequest](t, `{}`)
	if err == nil {
		t.Fatal("DecodeJSON({}) = nil, want une erreur de validation")
	}

	if want := "IssuerURL is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestDecodeJSONFallsBackToTheGoFieldOnDashTag(t *testing.T) {
	t.Parallel()

	_, err := decode[skippedRequest](t, `{}`)
	if err == nil {
		t.Fatal("DecodeJSON({}) = nil, want une erreur de validation")
	}

	if want := "Internal is required"; err.Error() != want {
		t.Errorf("err = %q, want %q", err.Error(), want)
	}
}

func TestDecodeJSONValidBody(t *testing.T) {
	t.Parallel()

	got, err := decode[taggedRequest](t, `{"issuerUrl":"https://idp.example.com","clientSecret":"averylongsecret"}`)
	if err != nil {
		t.Fatalf("DecodeJSON = %v, want nil", err)
	}

	if got.IssuerURL != "https://idp.example.com" {
		t.Errorf("IssuerURL = %q, want %q", got.IssuerURL, "https://idp.example.com")
	}
}

func TestDecodeJSONRejectsMalformedBody(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"json invalide": `{`,
		"champ inconnu": `{"issuerUrl":"https://idp.example.com","clientSecret":"averylongsecret","nope":1}`,
		"corps vide":    ``,
	}

	for name, body := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := decode[taggedRequest](t, body)
			if err == nil {
				t.Fatalf("DecodeJSON(%q) = nil, want une erreur", body)
			}

			if want := "invalid request body"; err.Error() != want {
				t.Errorf("err = %q, want %q", err.Error(), want)
			}
		})
	}
}

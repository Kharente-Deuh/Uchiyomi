// SPDX-License-Identifier: AGPL-3.0-or-later

package pgmodels_test

import (
	"reflect"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/repository/pgmodels"
)

const (
	testEmail      = "bob@example.com"
	testEmailClaim = "email"
)

func TestClaimsValue(t *testing.T) {
	t.Parallel()

	t.Run("nil sérialise un objet vide", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims

		got, err := c.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		b, ok := got.([]byte)
		if !ok {
			t.Fatalf("Value() = %T, want []byte", got)
		}

		if string(b) != "{}" {
			t.Errorf("Value() = %q, want %q", b, "{}")
		}
	})

	t.Run("map sérialisée en JSON", func(t *testing.T) {
		t.Parallel()

		got, err := pgmodels.Claims{testEmailClaim: testEmail}.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if want := `{"email":"bob@example.com"}`; string(got.([]byte)) != want {
			t.Errorf("Value() = %s, want %s", got, want)
		}
	})

	t.Run("map vide", func(t *testing.T) {
		t.Parallel()

		got, err := pgmodels.Claims{}.Value()
		if err != nil {
			t.Fatalf("Value: %v", err)
		}

		if string(got.([]byte)) != "{}" {
			t.Errorf("Value() = %s, want {}", got)
		}
	})
}

func TestClaimsScan(t *testing.T) {
	t.Parallel()

	t.Run("nil remet la map à nil", func(t *testing.T) {
		t.Parallel()

		c := pgmodels.Claims{"stale": true}
		if err := c.Scan(nil); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		if c != nil {
			t.Errorf("Scan(nil) = %v, want nil", c)
		}
	})

	t.Run("JSON valide", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims
		if err := c.Scan([]byte(`{"email":"bob@example.com","admin":true}`)); err != nil {
			t.Fatalf("Scan: %v", err)
		}

		want := pgmodels.Claims{testEmailClaim: testEmail, "admin": true}
		if !reflect.DeepEqual(c, want) {
			t.Errorf("Scan() = %v, want %v", c, want)
		}
	})

	t.Run("JSON invalide", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims
		if err := c.Scan([]byte(`{`)); err == nil {
			t.Error("Scan sur un JSON tronqué = nil, want une erreur")
		}
	})

	t.Run("type inattendu", func(t *testing.T) {
		t.Parallel()

		var c pgmodels.Claims

		err := c.Scan(`{"email":"bob@example.com"}`)
		if err == nil {
			t.Fatal("Scan(string) = nil, want une erreur")
		}

		if want := "claims: type string inattendu"; err.Error() != want {
			t.Errorf("err = %q, want %q", err.Error(), want)
		}
	})
}

func TestClaimsRoundTrip(t *testing.T) {
	t.Parallel()

	original := pgmodels.Claims{testEmailClaim: testEmail, "groups": []any{"admins", "users"}}

	encoded, err := original.Value()
	if err != nil {
		t.Fatalf("Value: %v", err)
	}

	var decoded pgmodels.Claims
	if err := decoded.Scan(encoded); err != nil {
		t.Fatalf("Scan: %v", err)
	}

	if !reflect.DeepEqual(decoded, original) {
		t.Errorf("aller-retour = %v, want %v", decoded, original)
	}
}

func TestGormDataType(t *testing.T) {
	t.Parallel()

	if got := (pgmodels.Claims{}).GormDataType(); got != "jsonb" {
		t.Errorf("GormDataType() = %q, want %q", got, "jsonb")
	}
}

func TestTableNames(t *testing.T) {
	t.Parallel()

	tests := map[string]string{
		"user":               pgmodels.User{}.TableName(),
		"password creds":     pgmodels.PasswordCreds{}.TableName(),
		"oidc provider":      pgmodels.OIDCProvider{}.TableName(),
		"federated identity": pgmodels.FederatedIdentity{}.TableName(),
	}

	want := map[string]string{
		"user":               "users",
		"password creds":     "password_credentials",
		"oidc provider":      "oidc_providers",
		"federated identity": "federated_identities",
	}

	for model, got := range tests {
		if got != want[model] {
			t.Errorf("%s: TableName() = %q, want %q", model, got, want[model])
		}
	}
}

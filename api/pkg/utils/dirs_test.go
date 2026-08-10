// SPDX-License-Identifier: AGPL-3.0-or-later

package utils_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kharente-deuh/uchiyomi-server/pkg/utils"
)

func TestEnsureDirCreatesUsableDirectory(t *testing.T) {
	t.Parallel()

	dir := filepath.Join(t.TempDir(), "cache", "images")

	if err := utils.EnsureDir(dir); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("os.Stat: %v", err)
	}

	if !info.IsDir() {
		t.Fatalf("%s is not a directory", dir)
	}

	if err := os.WriteFile(filepath.Join(dir, "probe.txt"), []byte("ok"), 0o600); err != nil {
		t.Fatalf("directory created but unusable (mode %v): %v", info.Mode().Perm(), err)
	}
}

func TestEnsureDirIsIdempotent(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()

	for i := range 3 {
		if err := utils.EnsureDir(dir); err != nil {
			t.Fatalf("EnsureDir call %d: %v", i+1, err)
		}
	}
}

func TestEnsureDirFailsOnExistingFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(path, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	if err := utils.EnsureDir(path); err == nil {
		t.Error("EnsureDir must fail when path is a file")
	}
}

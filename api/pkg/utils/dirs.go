// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"log/slog"
	"os"
	"syscall"
)

func EnsureDir(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err == nil {
		return ensureDirWritable(path)
	}

	if os.IsExist(err) {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("os.Stat: %w", err)
		}

		if !info.IsDir() {
			return fmt.Errorf("path %s exists but is not a directory", path)
		}

		return ensureDirWritable(path)
	}

	return fmt.Errorf("os.MkdirAll %s: %w", path, err)
}

func ensureDirWritable(path string) error {
	probe, err := os.CreateTemp(path, ".write-probe-*")
	if err != nil {
		return fmt.Errorf("directory %s is not writable: %w", path, err)
	}

	probeName := probe.Name()
	if err := probe.Close(); err != nil {
		return fmt.Errorf("probe.Close %s: %w", probeName, err)
	}

	if err := os.Remove(probeName); err != nil {
		return fmt.Errorf("os.Remove %s: %w", probeName, err)
	}

	return nil
}

func PrepareDataDir(logger *slog.Logger, path string, uid, gid int) (string, error) {
	if os.Geteuid() == 0 {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return "", fmt.Errorf("os.MkdirAll %s: %w", path, err)
		}

		if err := os.Chown(path, uid, gid); err != nil {
			return "", fmt.Errorf("os.Chown %s: %w", path, err)
		}

		if err := dropPrivileges(uid, gid); err != nil {
			return "", fmt.Errorf("dropPrivileges: %w", err)
		}

		if logger != nil {
			logger.Info("dropped privileges after data directory setup", "uid", uid, "gid", gid)
		}
	}

	if err := EnsureDir(path); err != nil {
		return "", err
	}

	if logger != nil {
		logger.Debug("cache directory ready", "dir", path)
	}

	return path, nil
}

func dropPrivileges(uid, gid int) error {
	if err := syscall.Setgid(gid); err != nil {
		return fmt.Errorf("syscall.Setgid(%d): %w", gid, err)
	}

	if err := syscall.Setuid(uid); err != nil {
		return fmt.Errorf("syscall.Setuid(%d): %w", uid, err)
	}

	return nil
}

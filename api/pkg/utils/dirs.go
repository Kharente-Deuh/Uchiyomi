// SPDX-License-Identifier: AGPL-3.0-or-later

package utils

import (
	"fmt"
	"os"
)

func EnsureDir(path string) error {
	err := os.MkdirAll(path, 0o755)
	if err == nil {
		return nil
	}

	if os.IsExist(err) {
		info, err := os.Stat(path)
		if err != nil {
			return fmt.Errorf("os.Stat: %w", err)
		}

		if !info.IsDir() {
			return fmt.Errorf("path %s exists but is not a directory", path)
		}

		return nil
	}

	return fmt.Errorf("os.MkdirAll %s: %w", path, err)
}

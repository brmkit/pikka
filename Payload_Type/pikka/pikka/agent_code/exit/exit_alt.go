//go:build !windows

package exit

import (
	"fmt"
	"os"
	"path/filepath"
)

func selfDelete() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return fmt.Errorf("failed to resolve executable path: %w", err)
	}

	if err := os.Remove(exePath); err != nil {
		return fmt.Errorf("failed to delete executable: %w", err)
	}
	return nil
}

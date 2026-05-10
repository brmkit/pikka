//go:build !windows

package exit

import (
	"os"
)

func selfDelete() error {
	exePath, err := os.Executable()
	if err != nil {
		return err
	}
	return os.Remove(exePath)
}

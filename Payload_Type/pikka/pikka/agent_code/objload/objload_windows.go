//go:build windows

package objload

import (
	"fmt"

	"github.com/praetorian-inc/goffloader/src/coff"
	"github.com/praetorian-inc/goffloader/src/lighthouse"
	"github.com/praetorian-inc/goffloader/src/pe"
)

func executeObj(fileBytes []byte, mode string, args []string, entryPoint string) (string, error) {
	switch mode {
	case "bof":
		var argBytes []byte
		var err error

		if len(args) > 0 {
			argBytes, err = lighthouse.PackArgs(args)
			if err != nil {
				return "", fmt.Errorf("failed to pack BOF arguments: %s", err.Error())
			}
		}

		return coff.LoadWithMethod(fileBytes, argBytes, entryPoint)

	case "pe":
		return pe.RunExecutable(fileBytes, args)

	default:
		return "", fmt.Errorf("unsupported mode: %s (use 'bof' or 'pe')", mode)
	}
}

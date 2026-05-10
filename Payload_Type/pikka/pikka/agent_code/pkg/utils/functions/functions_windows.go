//go:build windows

package functions

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

func isElevated() bool {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false
	}
	defer token.Close()

	return token.IsElevated()
}

func getArchitecture() string {
	return runtime.GOARCH
}

func getProcessName() string {
	name, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Base(name)
}

func getDomain() string {
	return os.Getenv("USERDOMAIN")
}

func getOS() string {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, "SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", registry.QUERY_VALUE)
	if err != nil {
		return "Windows (Unknown Version)"
	}
	defer k.Close()

	productName, _, err := k.GetStringValue("ProductName")
	if err != nil {
		return "Windows"
	}

	currentBuild, _, err := k.GetStringValue("CurrentBuild")
	if err == nil {
		return fmt.Sprintf("%s (Build %s)", productName, currentBuild)
	}
	return productName
}

func getUser() string {
	u, err := user.Current()
	if err != nil {
		return ""
	}
	return u.Username
}

func getEffectiveUser() string {
	return getUser()
}

func getPID() int {
	return os.Getpid()
}

func getHostname() string {
	h, err := os.Hostname()
	if err != nil {
		return ""
	}
	return h
}

func getCwd() string {
	d, err := os.Getwd()
	if err != nil {
		return ""
	}
	return d
}

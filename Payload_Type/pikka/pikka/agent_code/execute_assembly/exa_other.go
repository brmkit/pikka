//go:build !windows

package execute_assembly

// Standard

func executeAssembly(exebytes []byte, args []string) (string, string) {
	// not implemented on non-windows systems
	stdout := ""
	stderr := "execute_assembly is only supported on Windows targets"
	return stdout, stderr
}

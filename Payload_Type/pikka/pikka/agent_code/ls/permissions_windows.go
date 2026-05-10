//go:build windows

package ls

import (
	"os"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
)

func GetPermission(finfo os.FileInfo) structs.FilePermission {
	perms := structs.FilePermission{}
	perms.Permissions = finfo.Mode().Perm().String()
	if finfo.IsDir() {
		perms.Permissions = "d" + perms.Permissions[1:]
	}
	return perms
}

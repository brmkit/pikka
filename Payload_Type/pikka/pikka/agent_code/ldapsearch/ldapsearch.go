package ldapsearch

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
	"github.com/go-ldap/ldap/v3"
)

func (e *LdapQuery) parseStringArray(configArray []interface{}) []string {
	urls := make([]string, len(configArray))
	for l, p := range configArray {
		urls[l] = p.(string)
	}
	return urls
}

type LdapQuery struct {
	Query     string
	Base      string
	Attrs     []string
	SizeLimit int
	// not windows
	Username string
	Password string
	Server   string
}

func (e *LdapQuery) UnmarshalJSON(data []byte) error {
	alias := map[string]interface{}{}
	err := json.Unmarshal(data, &alias)
	if err != nil {
		return err
	}
	if v, ok := alias["query"]; ok && v != nil {
		if s, ok := v.(string); ok {
			e.Query = s
		}
	}
	if v, ok := alias["base"]; ok && v != nil {
		if s, ok := v.(string); ok {
			e.Base = s
		}
	}
	if v, ok := alias["attributes"]; ok && v != nil {
		if arr, ok := v.([]interface{}); ok {
			e.Attrs = e.parseStringArray(arr)
		}
	}
	if v, ok := alias["limit"]; ok && v != nil {
		if f, ok := v.(float64); ok {
			e.SizeLimit = int(f)
		}
	}
	if v, ok := alias["username"]; ok && v != nil {
		if s, ok := v.(string); ok {
			e.Username = s
		}
	}
	if v, ok := alias["password"]; ok && v != nil {
		if s, ok := v.(string); ok {
			e.Password = s
		}
	}
	return nil
}

func Run(task structs.Task) {
	msg := task.NewResponse()

	var args LdapQuery

	if err := json.Unmarshal([]byte(task.Params), &args); err != nil {
		msg.SetError(fmt.Sprintf("Failed to unmarshal parameters: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}

	if args.Query == "" {
		// force default query
		args.Query = "(&(objectclass=top)(objectclass=container))"
	}

	if args.SizeLimit <= 0 {
		args.SizeLimit = 0 // default limit
	}

	// extract dc information from environment
	server, err := resolveLDAPServer()
	if err != nil {
		msg.SetError(fmt.Sprintf("Failed to resolve LDAP server: %s", err.Error()))
		task.Job.SendResponses <- msg
		return
	}

	args.Server = server

	conn, err := ldap.DialURL("ldap://" + server)
	if err != nil {
		msg.SetError(err.Error())
		task.Job.SendResponses <- msg
		return
	}
	defer conn.Close()

	if err := ldapBind(conn, &args); err != nil {
		msg.SetError(fmt.Sprintf("bind failed: %v", err))
		task.Job.SendResponses <- msg
		return
	}

	scope := ldap.ScopeWholeSubtree
	if args.Base == "" {
		scope = ldap.ScopeBaseObject
	}

	req := ldap.NewSearchRequest(
		args.Base,
		scope,
		ldap.NeverDerefAliases,
		args.SizeLimit,
		0,
		false,
		args.Query,
		args.Attrs,
		nil,
	)

	res, err := conn.Search(req)

	if err != nil {
		msg.SetError(fmt.Sprintf("Error: %s", err))
		task.Job.SendResponses <- msg
		return
	}

	browserEntries := make([]structs.CustomBrowserEntry, 0, len(res.Entries))

	for _, ldapEntry := range res.Entries {
		distinguishedName := ldapEntry.DN
		// relativeDN is the first part of the DN
		relativeDN := strings.Split(distinguishedName, ",")[0]
		// build CustomBrowser entry from the LDAP result
		browserEntry := structs.CustomBrowserEntry{
			Name:            relativeDN,
			DisplayPath:     distinguishedName,
			ParentPath:      parentPathReversed(distinguishedName),
			CanHaveChildren: canHaveChildren(ldapEntry),
			Metadata:        buildMetadata(ldapEntry),
		}
		browserEntries = append(browserEntries, browserEntry)
	}

	msg.Completed = true
	msg.CustomBrowser = &structs.CustomBrowser{
		BrowserName:     "ldap_browser",
		Host:            args.Base,
		SetAsUserOutput: true,
		Entries:         browserEntries,
	}

	task.Job.SendResponses <- msg

}

package ldapsearch

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/MythicAgents/pikka/Payload_Type/pikka/agent_code/pkg/utils/structs"
	"github.com/go-ldap/ldap/v3"
)

type LdapQuery struct {
	Query      string   `json:"query"`
	Base       string   `json:"base"`
	Attrs      []string `json:"attributes"`
	SizeLimit  int      `json:"limit"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	Server     string   `json:"server"`
	UseTLS     bool     `json:"use_tls"`
	SkipVerify bool     `json:"skip_verify"`
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

	if args.Server == "" {
		server, err := resolveLDAPServer()
		if err != nil {
			// fallback: derive domain from base DN (e.g. "DC=red,DC=local" -> "red.local")
			if domain := domainFromBaseDN(args.Base); domain != "" {
				if dc, dcErr := detectDC(domain); dcErr == nil && dc != "" {
					server = dc
				} else {
					server = domain
				}
			}
			if server == "" {
				msg.SetError(fmt.Sprintf("Failed to resolve LDAP server: %s", err.Error()))
				task.Job.SendResponses <- msg
				return
			}
		}
		args.Server = server
	}

	var conn *ldap.Conn
	var connErr error
	dialer := net.Dialer{Timeout: 10 * time.Second}

	tlsConfig := &tls.Config{
		ServerName:         args.Server,
		InsecureSkipVerify: args.SkipVerify,
	}

	if args.UseTLS {
		conn, connErr = ldap.DialURL("ldaps://"+args.Server, ldap.DialWithDialer(&dialer), ldap.DialWithTLSConfig(tlsConfig))
	} else {
		conn, connErr = ldap.DialURL("ldap://"+args.Server, ldap.DialWithDialer(&dialer))
		if connErr == nil {
			connErr = conn.StartTLS(tlsConfig)
		}
	}
	if connErr != nil {
		msg.SetError(connErr.Error())
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

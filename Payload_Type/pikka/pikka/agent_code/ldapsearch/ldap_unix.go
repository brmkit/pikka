//go:build !windows

package ldapsearch

import (
	"fmt"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// Limitation: Linux only supports simple bind with explicit credentials.
// Kerberos/GSSAPI bind is available on Windows via SSPI.
func ldapBind(conn *ldap.Conn, args *LdapQuery) error {
	if args.Username == "" || args.Password == "" {
		return fmt.Errorf("explicit credentials required on linux")
	}

	return conn.Bind(args.Username, args.Password)
}

func resolveLDAPServer() (string, error) {
	// check if server is explicitly set
	if v := os.Getenv("LDAP_SERVER"); v != "" {
		return strings.ToLower(v), nil
	}

	// try to detect domain and resolve DC via SRV records
	domainKeys := []string{"KRB5_REALM", "USERDNSDOMAIN"}
	for _, key := range domainKeys {
		if domain := os.Getenv(key); domain != "" {
			if dc, err := detectDC(domain); err == nil && dc != "" {
				return dc, nil
			}
		}
	}

	return "", fmt.Errorf("no LDAP server context available")
}

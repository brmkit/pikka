//go:build !windows

package ldapsearch

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

func ldapBind(conn *ldap.Conn, args *LdapQuery) error {
	if args.Username == "" || args.Password == "" {
		return fmt.Errorf("explicit credentials required on linux")
	}

	return conn.Bind(args.Username, args.Password)
}

func resolveLDAPServer() (string, error) {
	if v := os.Getenv("LDAP_SERVER"); v != "" {
		return strings.ToLower(v), nil
	}

	domainKeys := []string{"KRB5_REALM", "USERDNSDOMAIN"}
	for _, key := range domainKeys {
		if domain := os.Getenv(key); domain != "" {
			if dc, err := detectDC(domain); err == nil && dc != "" {
				return dc, nil
			}
		}
	}

	// fallback: parse default_realm from /etc/krb5.conf
	if realm := krb5DefaultRealm(); realm != "" {
		if dc, err := detectDC(realm); err == nil && dc != "" {
			return dc, nil
		}
	}

	return "", fmt.Errorf("no LDAP server context available")
}

func krb5DefaultRealm() string {
	f, err := os.Open("/etc/krb5.conf")
	if err != nil {
		return ""
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "default_realm") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}
	return ""
}

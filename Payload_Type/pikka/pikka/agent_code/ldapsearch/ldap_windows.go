//go:build windows

package ldapsearch

import (
	"fmt"
	"os"
	"strings"

	"github.com/alexbrainman/sspi"
	"github.com/alexbrainman/sspi/kerberos"
	"github.com/go-ldap/ldap/v3"
)

// mostly derived from https://github.com/go-ldap/ldap/tree/master/v3/gssapi

// SSPIClient implements ldap.GSSAPIClient
type SSPIClient struct {
	creds *sspi.Credentials
	ctx   *kerberos.ClientContext
}

func NewSSPIClient() (*SSPIClient, error) {
	creds, err := kerberos.AcquireCurrentUserCredentials()
	if err != nil {
		return nil, err
	}
	return &SSPIClient{creds: creds}, nil
}

func (c *SSPIClient) DeleteSecContext() error {
	if c.ctx != nil {
		return c.ctx.Release()
	}
	return nil
}

func (c *SSPIClient) Close() error {
	_ = c.DeleteSecContext()
	return c.creds.Release()
}

func (c *SSPIClient) InitSecContext(target string, token []byte) ([]byte, bool, error) {
	return c.InitSecContextWithOptions(target, token, nil)
}

func (c *SSPIClient) InitSecContextWithOptions(target string, token []byte, opts []int) ([]byte, bool, error) {
	flags := uint32(sspi.ISC_REQ_INTEGRITY | sspi.ISC_REQ_CONFIDENTIALITY | sspi.ISC_REQ_MUTUAL_AUTH)

	if token == nil {
		ctx, completed, out, err := kerberos.NewClientContextWithFlags(c.creds, target, flags)
		if err != nil {
			return nil, false, err
		}
		c.ctx = ctx
		return out, !completed, nil
	}

	completed, out, err := c.ctx.Update(token)
	if err != nil {
		return nil, false, err
	}
	if err := c.ctx.VerifyFlags(); err != nil {
		return nil, false, err
	}
	return out, !completed, nil
}

func (c *SSPIClient) NegotiateSaslAuth(token []byte, authzid string) ([]byte, error) {
	// Unwrap server token
	_, payload, err := c.ctx.DecryptMessage(token, 0)
	if err != nil {
		return nil, err
	}

	// RFC 4752 §3.1: MUST be exactly 4 bytes
	if len(payload) != 4 {
		return nil, fmt.Errorf("invalid SASL token length: %d", len(payload))
	}

	// Select: no security layer, max size 0
	response := []byte{0x00, 0x00, 0x00, 0x00}

	// Wrap response (SSPI decides sign vs encrypt)
	return c.ctx.EncryptMessage(response, 0, 0)
}

func resolveLDAPServer() (string, error) {
	logon := strings.TrimLeft(os.Getenv("LOGONSERVER"), "\\")
	domain := os.Getenv("USERDNSDOMAIN")

	if logon == "" || domain == "" {
		return "", fmt.Errorf("LOGONSERVER or USERDNSDOMAIN not set")
	}

	if dc, err := detectDC(domain); err == nil && dc != "" {
		return dc, nil
	}

	return "", fmt.Errorf("no LDAP server context available")
}

func ldapBind(conn *ldap.Conn, args *LdapQuery) error {

	// set SPN
	spn := "ldap/" + args.Server

	client, err := NewSSPIClient()
	if err != nil {
		return err
	}
	defer client.Close()

	// use process security context
	return conn.GSSAPIBind(client, spn, "")
}

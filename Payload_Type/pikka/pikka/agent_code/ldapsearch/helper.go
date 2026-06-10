package ldapsearch

import (
	"encoding/binary"
	"fmt"
	"net"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// TODO: when manually browsing the object list the path is reversed
// but when querying by clicking on the right panel the path is correct

func convertLDAPValue(attrName string, values [][]byte) interface{} {
	if len(values) == 0 {
		return nil
	}

	if len(values) == 1 {
		return convertSingleValue(attrName, values[0])
	}

	out := make([]interface{}, 0, len(values))
	for _, v := range values {
		out = append(out, convertSingleValue(attrName, v))
	}
	return out
}

func convertSingleValue(attrName string, value []byte) interface{} {
	switch strings.ToLower(attrName) {

	case "objectguid":
		guid, err := ldapObjectGUIDToString(value)
		if err != nil {
			return nil
		}
		return guid

	case "objectsid":
		sid, err := ldapSIDToString(value)
		if err != nil {
			return nil
		}
		return sid

	default:
		return string(value)
	}
}

func ldapObjectGUIDToString(b []byte) (string, error) {
	if len(b) != 16 {
		return "", fmt.Errorf("invalid objectGUID length")
	}

	return fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[3], b[2], b[1], b[0],
		b[5], b[4],
		b[7], b[6],
		b[8], b[9],
		b[10], b[11], b[12], b[13], b[14], b[15],
	), nil
}

func ldapSIDToString(b []byte) (string, error) {
	if len(b) < 8 {
		return "", fmt.Errorf("invalid SID length")
	}

	revision := b[0]
	subAuthCount := int(b[1])

	authority := binary.BigEndian.Uint64(append([]byte{0, 0}, b[2:8]...))

	sid := fmt.Sprintf("S-%d-%d", revision, authority)

	offset := 8
	for i := 0; i < subAuthCount; i++ {
		if offset+4 > len(b) {
			return "", fmt.Errorf("invalid SID subauthority")
		}
		subAuth := binary.LittleEndian.Uint32(b[offset : offset+4])
		sid += fmt.Sprintf("-%d", subAuth)
		offset += 4
	}

	return sid, nil
}

func parentPathReversed(dn string) string {
	parts := strings.Split(dn, ",")
	if len(parts) <= 1 {
		return ""
	}

	// remove first element
	parts = parts[1:]

	// reverse
	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ",")
}

var groupClasses = map[string]struct{}{
	"group":              {},
	"domain":             {},
	"organizationalunit": {},
	"container":          {},
}

func canHaveChildren(e *ldap.Entry) bool {
	classes := e.GetAttributeValues("objectClass")
	for _, v := range classes {
		if _, ok := groupClasses[strings.ToLower(v)]; ok {
			return true
		}
	}
	return false
}

func buildMetadata(e *ldap.Entry) map[string]interface{} {
	meta := make(map[string]interface{})

	for _, attr := range e.Attributes {
		meta[attr.Name] = convertLDAPValue(attr.Name, attr.ByteValues)
	}

	return meta
}

func getBaseDN(domain string) (string, error) {
	if domain == "" {
		return "", fmt.Errorf("invalid input")
	}

	dcs := "DC=" + strings.ReplaceAll(domain, ".", ",DC=")

	return strings.ToUpper(dcs), nil
}

// TODO: verify this function with gohawk
func detectDC(domain string) (string, error) {
	// input: domain name
	// output: DC server FQDN

	// query _ldap._tcp.dc._msdcs.<domain>
	_, addrs, err := net.LookupSRV("ldap", "tcp", "dc._msdcs."+domain)
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("no SRV records found for domain %s", domain)
	}

	dc := strings.TrimSuffix(addrs[0].Target, ".")
	return strings.ToLower(dc), nil
}

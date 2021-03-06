package config

import (
	"net"
	"os"
	"strings"
        "github.com/pkg/errors"
)

// Get Fully Qualified Domain Name
// returns the host FQDN or an error
// Original idea from: https://github.com/Showmax/go-fqdn/blob/master/fqdn.go
func GetFQDN() (string, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return "unknown", errors.Wrapf(err, "hostname unknown")
	}

	addrs, err := net.LookupIP(hostname)
	if err != nil {
		return hostname, nil
	}

	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ip, err := ipv4.MarshalText()
			if err != nil {
				return hostname, nil
			}
			hosts, err := net.LookupAddr(string(ip))
			if err != nil || len(hosts) == 0 {
				return hostname, nil
			}
			fqdn := hosts[0]
			return strings.TrimSuffix(fqdn, "."), nil // return fqdn without trailing dot
		}
	}
	return hostname, nil
}

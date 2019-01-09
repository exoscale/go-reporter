package config

import (
	"net"

	"github.com/pkg/errors"
)

// Addr represents an endpoint (eg. 127.0.0.1:1654).
type Addr string

func (a Addr) String() string {
	return string(a)
}

// UnmarshalText parses text as an address of the form "host:port" or
// "[ipv6-host%zone]:port" and resolves a pair of domain name and port
// name. A literal address or host name for IPv6 must be enclosed in
// square brackets, as in "[::1]:80", "[ipv6-host]:http" or
// "[ipv6-host%zone]:80". Resolution of names is biased to IPv4.
func (a *Addr) UnmarshalText(text []byte) error {
	rawAddr := string(text)
	addr, err := net.ResolveTCPAddr("tcp", rawAddr)
	if err != nil {
		return errors.Wrapf(err, "unable to solve %q", rawAddr)
	}
	*a = Addr(addr.String())
	return nil
}

func (hostname Hostname) String() string {
	return net.IP(hostname).String()
}

// Hostname is any valid hostname that will resolve to one IPv4
// address. Can be an IP address or a valid and resolvable hostname.
type Hostname net.IP

// UnmarshalText parses and validates a Hostname.
func (hostname *Hostname) UnmarshalText(text []byte) error {
	rawHostname := string(text)
	addrs, err := net.LookupIP(rawHostname)
	if err != nil {
		return errors.Wrapf(err, "unable to solve %q", rawHostname)
	}
	if len(addrs) == 0 {
		return errors.Errorf("no IP address for %q", rawHostname)
	}

	*hostname = make(Hostname, len(addrs[0]))
	copy(*hostname, addrs[0])
	return nil
}

// MACAddr is any valid mac address
type MACAddr net.HardwareAddr

func (mac MACAddr) String() string {
	return net.HardwareAddr(mac).String()
}

// UnmarshalText parses and validates a MACAddr.
func (mac *MACAddr) UnmarshalText(text []byte) error {
	rawMac := string(text)
	addr, err := net.ParseMAC(rawMac)
	if err != nil {
		return errors.Wrapf(err, "unable to parse %s as a mac addr", rawMac)
	}
	*mac = MACAddr(addr)
	return nil
}

package helpers

import (
	"encoding/binary"
	"net"
)

// IncrementIPv4 will increment the IP address from the given number.
// Increment can be positive or negative.
func IncrementIPv4(ip net.IP, inc int) net.IP {
	ip = ip.To4()
	v := binary.BigEndian.Uint32(ip)
	if v >= uint32(0) {
		v = v + uint32(inc)
	} else {
		v = v - uint32(-inc)
	}
	ip = make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, v)
	return ip
}

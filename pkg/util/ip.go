package util

import (
	"net"
	"strconv"

	"github.com/miekg/dns"
)

func GetMask(ip net.IP) net.IPMask {
	if ip.To4() != nil {
		return net.CIDRMask(32, 32)
	}
	return net.CIDRMask(128, 128)
}

func GetServer(config dns.ClientConfig) string {
	var port int
	if port, _ = strconv.Atoi(config.Port); port == 0 {
		port = 53
	}
	var server string
	for _, s := range config.Servers {
		ip := net.ParseIP(s)
		if ip.IsLoopback() || ip.IsUnspecified() {
			continue
		}
		server = s
		break
	}
	return net.JoinHostPort(server, strconv.Itoa(port))
}

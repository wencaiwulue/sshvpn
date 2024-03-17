package util

import (
	"net"
	"strconv"

	"github.com/miekg/dns"
	"k8s.io/apimachinery/pkg/util/sets"
)

func GetMask(ip net.IP) net.IPMask {
	if ip.To4() != nil {
		return net.CIDRMask(32, 32)
	}
	return net.CIDRMask(128, 128)
}

func GetServer(config dns.ClientConfig, dnsConfig *dns.ClientConfig) string {
	var port int
	if port, _ = strconv.Atoi(config.Port); port == 0 {
		port = 53
	}
	var set = sets.New[string](dnsConfig.Servers...)
	var server string
	for _, s := range config.Servers {
		ip := net.ParseIP(s)
		if ip.IsLoopback() || ip.IsUnspecified() {
			continue
		}
		if set.Has(s) {
			continue
		}
		server = s
		break
	}
	return net.JoinHostPort(server, strconv.Itoa(port))
}

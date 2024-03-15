package config

import (
	"net"

	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

const (
	v4 = "223.253.0.1"
	v6 = "efff:ffff:ffff:ffff:ffff:ffff:ffff:8888"
)

var (
	ipv4  = net.ParseIP(v4)
	ipv6  = net.ParseIP(v6)
	Addr4 = (&net.IPNet{IP: ipv4, Mask: util.GetMask(ipv4)}).String()
	Addr6 = (&net.IPNet{IP: ipv6, Mask: util.GetMask(ipv6)}).String()
)

type ProxyMode string

const (
	ProxyModeFull  = ProxyMode("full")
	ProxyModeSplit = ProxyMode("split")
)

var (
	TCPPort = 56788
	UDPPort = 56789
)

type StackType string

const (
	SingleStackIPv4 = StackType("ipv4")
	SingleStackIPv6 = StackType("ipv6")
	DualStack       = StackType("dual")
)

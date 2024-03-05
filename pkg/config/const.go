package config

type ProxyType string

const (
	ProxyTypePAC    = ProxyType("pac")
	ProxyTypeGlobal = ProxyType("global")
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

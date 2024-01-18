package config

type ProxyType string

const (
	ProxyTypePAC   = ProxyType("pac")
	ProxyTypeGlobe = ProxyType("globe")
)

var (
	TCPPort = 56788
	UDPPort = 56789
)

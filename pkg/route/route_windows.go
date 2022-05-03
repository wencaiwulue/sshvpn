package route

import (
	"net"
	"strconv"

	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

func addIP(ifName string, mtu uint64, ipNet net.IPNet) error {
	formatUint, err := strconv.ParseUint(ifName, 0, 64)
	if err != nil {
		return err
	}
	luid := winipcfg.LUID(formatUint)
	return luid.AddIPAddress(ipNet)
}

func addTunRoutes(ifName string, routes ...*net.IPNet) error {
	formatUint, err := strconv.ParseUint(ifName, 0, 64)
	if err != nil {
		return err
	}
	luid := winipcfg.LUID(formatUint)
	_ = luid.FlushRoutes(windows.AF_INET)
	for _, route := range routes {
		if route == nil {
			continue
		}
		if err = luid.AddRoute(route, net.IPv4(0, 0, 0, 0), 0); err != nil {
			return err
		}
	}
	return nil
}

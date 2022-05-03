package route

import (
	"net"
	"strconv"

	"golang.zx2c4.com/wireguard/tun"
)

func AddRoute(t tun.Device, route []*net.IPNet) error {
	// if ok, Windows
	if obj, ok := t.(interface{ LUID() uint64 }); ok {
		luid := obj.LUID()
		return addTunRoutes(strconv.FormatUint(luid, 10), route...)
	}
	name, err := t.Name()
	if err != nil {
		return err
	}
	return addTunRoutes(name, route...)
}

func AddIP(device tun.Device, ipNet net.IPNet) error {
	mtu, err := device.MTU()
	if err != nil {
		return err
	}
	name, err := device.Name()
	if err != nil {
		return err
	}
	return addIP(name, mtu, ipNet)
}

func AddIPAndRoute(device tun.Device, ipNet net.IPNet, route []*net.IPNet) error {
	err := AddIP(device, ipNet)
	if err != nil {
		return err
	}
	return AddRoute(device, route)
}

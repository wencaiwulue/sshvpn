package route

import (
	"errors"
	"fmt"
	"net"
	"syscall"

	"github.com/docker/libcontainer/netlink"
	log "github.com/sirupsen/logrus"
)

func addIP(ifName string, mtu uint64, ipNet net.IPNet) error {
	cmd := fmt.Sprintf("ip link set dev %s mtu %d", ifName, mtu)
	log.Debug("[tun]", cmd)
	if err := netlink.NetworkLinkAddIp(ifce, ipNet.IP, ipNet); err != nil {
		return err
	}
	if err := netlink.NetworkLinkUp(ifce); err != nil {
		return err
	}
	if err := netlink.NetworkSetMTU(ifce, 1359); err != nil {
		return err
	}
}

func addTunRoutes(ifName string, routes ...*net.IPNet) error {
	for _, route := range routes {
		if route == nil {
			continue
		}
		cmd := fmt.Sprintf("ip route add %s dev %s", route.String(), ifName)
		log.Debugf("[tun] %s", cmd)
		if err := netlink.AddRoute(route.String(), "", "", ifName); err != nil && !errors.Is(err, syscall.EEXIST) {
			return fmt.Errorf("%s: %v", cmd, err)
		}
	}
	return nil
}

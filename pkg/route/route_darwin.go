package route

import (
	"fmt"
	"net"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

func addIP(ifName string, mtu int, ipNet net.IPNet) error {
	cmd := fmt.Sprintf("ifconfig %s inet %s %s mtu %d up", ifName, ipNet.String(), ipNet.IP.String(), mtu)
	log.Debug("[tun]", cmd)
	args := strings.Split(cmd, " ")
	if err := exec.Command(args[0], args[1:]...).Run(); err != nil {
		return fmt.Errorf("%s: %v", cmd, err)
	}
	return nil
}

func addTunRoutes(ifName string, routes ...*net.IPNet) error {
	for _, route := range routes {
		if route == nil {
			continue
		}
		cmd := fmt.Sprintf("route add -net %s -interface %s", route.String(), ifName)
		log.Debug("[tun]", cmd)
		args := strings.Split(cmd, " ")
		if er := exec.Command(args[0], args[1:]...).Run(); er != nil {
			return fmt.Errorf("%s: %v", cmd, er)
		}
	}
	return nil
}

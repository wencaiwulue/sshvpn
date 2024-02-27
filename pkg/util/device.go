package util

import (
	"net"
	"os/exec"
	"strings"

	"github.com/libp2p/go-netroute"
	"github.com/pkg/errors"
)

func GetDefaultDevice() (*string, error) {
	router, err := netroute.New()
	if err != nil {
		return nil, err
	}
	iface, _, _, err := router.Route(net.ParseIP("8.8.8.8"))
	if err != nil {
		return nil, err
	}
	return GetHardwarePort(iface.Name)
}

// GetHardwarePort
/*
*
➜  ~ networksetup -listallhardwareports

Hardware Port: USB 10/100/1000 LAN
Device: en7
Ethernet Address: 44:67:52:05:58:12

Hardware Port: Ethernet Adapter (en4)
Device: en4
Ethernet Address: 7a:91:c0:34:c1:b1

Hardware Port: Ethernet Adapter (en5)
Device: en5
Ethernet Address: 7a:91:c0:34:c1:b2

Hardware Port: Ethernet Adapter (en6)
Device: en6
Ethernet Address: 7a:91:c0:34:c1:b3

Hardware Port: Thunderbolt Bridge
Device: bridge0
Ethernet Address: 36:01:f7:e3:03:c0

Hardware Port: Wi-Fi
Device: en0
Ethernet Address: bc:d0:74:4c:97:90

Hardware Port: Thunderbolt 1
Device: en1
Ethernet Address: 36:01:f7:e3:03:c0

Hardware Port: Thunderbolt 2
Device: en2
Ethernet Address: 36:01:f7:e3:03:c4

Hardware Port: Thunderbolt 3
Device: en3
Ethernet Address: 36:01:f7:e3:03:c8

VLAN Configurations
===================
➜  ~
*/
func GetHardwarePort(name string) (*string, error) {
	cmd := exec.Command("networksetup", "-listallhardwareports")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, err
	}
	outputStr := string(output)
	lines := strings.Split(outputStr, "\n")
	for i := 0; i < len(lines); i++ {
		if strings.Contains(lines[i], "Hardware Port") && i+1 < len(lines) && strings.Contains(lines[i+1], name) {
			space := strings.TrimSpace(strings.Split(lines[i], ":")[1])
			return &space, nil
		}
	}
	return nil, errors.New("not found")
}

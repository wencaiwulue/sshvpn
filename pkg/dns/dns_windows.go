//go:build windows

package dns

import (
	"context"
	"net"
	"net/netip"

	"github.com/miekg/dns"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
	"golang.zx2c4.com/wireguard/windows/tunnel/winipcfg"
)

func GetDnsServers(ctx context.Context, tunDevice *net.Interface) (*dns.ClientConfig, error) {
	luid, err := winipcfg.LUIDFromIndex(uint32(tunDevice.Index))
	if err != nil {
		return nil, err
	}
	addrs, err := luid.DNS()
	if err != nil {
		return nil, err
	}
	var servers []string
	for _, addr := range addrs {
		servers = append(servers, net.IP(addr.AsSlice()).String())
	}
	config := dns.ClientConfig{
		Servers: servers,
	}
	return &config, nil
}

func SetDnsServers(ctx context.Context, config dns.ClientConfig, tunDevice *net.Interface) error {
	luid, err := winipcfg.LUIDFromIndex(uint32(tunDevice.Index))
	if err != nil {
		return err
	}
	var servers []netip.Addr
	for _, s := range config.Servers {
		var addr netip.Addr
		addr, err = netip.ParseAddr(s)
		if err != nil {
			log.Errorf("parse %s failed: %s", s, err)
			return err
		}
		servers = append(servers, addr.Unmap())
	}
	err = luid.SetDNS(windows.AF_INET, servers, config.Search)
	if err != nil {
		log.Errorf("set DNS failed: %s", err)
		return err
	}
	err = luid.SetDNS(windows.AF_INET6, servers, config.Search)
	if err != nil {
		log.Errorf("set DNS failed: %s", err)
		return err
	}
	return nil
}

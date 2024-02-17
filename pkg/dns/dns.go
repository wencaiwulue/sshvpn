package dns

import (
	"context"
	"net"

	"github.com/miekg/dns"
)

func Append(ctx context.Context, config dns.ClientConfig, device *net.Interface) error {
	local, err := GetDnsServers(ctx, device)
	if err != nil {
		return err
	}
	config.Servers = append(config.Servers, local.Servers...)
	err = SetDnsServers(ctx, config, device)
	return err
}

func Remove(ctx context.Context, config dns.ClientConfig, device *net.Interface) error {
	local, err := GetDnsServers(ctx, nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(local.Servers); i++ {
	out:
		for _, s := range config.Servers {
			if local.Servers[i] == s {
				local.Servers = append(local.Servers[:i], local.Servers[i+1:]...)
				i--
				continue out
			}
		}
	}
	config.Servers = local.Servers
	err = SetDnsServers(ctx, config, device)
	return err
}

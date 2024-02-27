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
	var find = func(servers []string, str string) bool {
		for _, server := range servers {
			if server == str {
				return true
			}
		}
		return false
	}
	var unique []string
	for _, server := range config.Servers {
		if !find(unique, server) {
			unique = append(unique, server)
		}
	}
	for _, server := range local.Servers {
		if !find(unique, server) {
			unique = append(unique, server)
		}
	}
	config.Servers = unique
	err = SetDnsServers(ctx, config, device)
	return err
}

func Remove(ctx context.Context, config dns.ClientConfig, device *net.Interface) error {
	local, err := GetDnsServers(ctx, device)
	if err != nil {
		return err
	}
	var findServer = func(server string) bool {
		for _, s := range config.Servers {
			if server == s {
				return true
			}
		}
		return false
	}
	for i := 0; i < len(local.Servers); i++ {
		if findServer(local.Servers[i]) {
			local.Servers = append(local.Servers[:i], local.Servers[i+1:]...)
			i--
		}
	}
	config.Servers = local.Servers
	err = SetDnsServers(ctx, config, device)
	return err
}

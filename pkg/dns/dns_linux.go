//go:build linux

package dns

import (
	"bytes"
	"context"
	"net"

	"github.com/docker/docker/libnetwork/resolvconf"
	"github.com/miekg/dns"
)

func GetDnsServers(ctx context.Context, device *net.Interface) (*dns.ClientConfig, error) {
	get, err := resolvconf.Get()
	if err != nil {
		return nil, err
	}
	resolvConf, err := dns.ClientConfigFromReader(bytes.NewBuffer(get.Content))
	if err != nil {
		return nil, err
	}
	return resolvConf, nil
}

func SetDnsServers(ctx context.Context, config dns.ClientConfig, device *net.Interface) error {
	get, err := resolvconf.Get()
	if err != nil {
		return err
	}
	resolvConf, err := dns.ClientConfigFromReader(bytes.NewBuffer(get.Content))
	if err != nil {
		return err
	}
	resolvConf.Servers = config.Servers
	options := resolvconf.GetOptions(get.Content)
	_, err = resolvconf.Build(resolvconf.Path(), resolvConf.Servers, resolvConf.Search, options)
	return err
}

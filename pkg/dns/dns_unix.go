//go:build darwin

package dns

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/docker/docker/libnetwork/resolvconf"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

// GetDnsServers networksetup -getdnsservers Wi-Fi
func GetDnsServers(ctx context.Context, device *net.Interface) (*dns.ClientConfig, error) {
	deviceName := "Wi-Fi"
	name, err := util.GetDefaultDevice()
	if err == nil {
		deviceName = *name
	}
	str := fmt.Sprintf("networksetup -getdnsservers %s", deviceName)
	log.Debug(str)
	split := strings.Split(str, " ")
	output, err := exec.CommandContext(ctx, split[0], split[1:]...).Output()
	if err != nil {
		return nil, err
	}
	var list []string
	i := strings.Split(string(output), "\n")
	for _, s := range i {
		s = strings.TrimSpace(s)
		if ip := net.ParseIP(s); ip != nil {
			list = append(list, s)
		}
	}
	if len(list) == 0 {
		var getDefaultNameServer = func() ([]string, error) {
			get, err := resolvconf.Get()
			if err != nil {
				return nil, err
			}
			resolvConf, err := dns.ClientConfigFromReader(bytes.NewBuffer(get.Content))
			if err != nil {
				return nil, err
			}
			return resolvConf.Servers, nil
		}
		if nameServer, _ := getDefaultNameServer(); nameServer != nil {
			list = append(list, nameServer...)
		}
	}
	config := dns.ClientConfig{
		Servers: list,
	}
	return &config, nil
}

// SetDnsServers networksetup -setdnsservers Wi-Fi 8.8.8.8 8.8.4.4
func SetDnsServers(ctx context.Context, config dns.ClientConfig, device *net.Interface) error {
	deviceName := "Wi-Fi"
	name, err := util.GetDefaultDevice()
	if err == nil {
		deviceName = *name
	}
	var str = fmt.Sprintf("networksetup -setdnsservers %s empty", deviceName)
	if len(config.Servers) != 0 {
		str = fmt.Sprintf("networksetup -setdnsservers %s %s", deviceName, strings.Join(config.Servers, " "))
	}
	log.Debug(str)
	split := strings.Split(str, " ")
	output, err := exec.CommandContext(ctx, split[0], split[1:]...).Output()
	if err != nil {
		return errors.Wrap(err, string(output))
	}
	return nil
}

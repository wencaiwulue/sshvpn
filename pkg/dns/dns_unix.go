//go:build darwin

package dns

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// GetDnsServers networksetup -getdnsservers Wi-Fi
func GetDnsServers(ctx context.Context, device *net.Interface) (*dns.ClientConfig, error) {
	str := "networksetup -getdnsservers Wi-Fi"
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
	config := dns.ClientConfig{
		Servers: list,
	}
	return &config, nil
}

// SetDnsServers networksetup -setdnsservers Wi-Fi 8.8.8.8 8.8.4.4
func SetDnsServers(ctx context.Context, config dns.ClientConfig, device *net.Interface) error {
	var str = fmt.Sprintf("networksetup -setdnsservers Wi-Fi empty")
	if len(config.Servers) != 0 {
		str = fmt.Sprintf("networksetup -setdnsservers Wi-Fi %s", strings.Join(config.Servers, " "))
	}
	log.Debug(str)
	split := strings.Split(str, " ")
	output, err := exec.CommandContext(ctx, split[0], split[1:]...).Output()
	if err != nil {
		return errors.Wrap(err, string(output))
	}
	return nil
}

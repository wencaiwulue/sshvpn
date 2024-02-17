//go:build darwin

package dns

import (
	"context"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

// networksetup -getdnsservers Wi-Fi
func GetDnsServers(ctx context.Context) ([]string, error) {
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
	return list, nil
}

// networksetup -setdnsservers Wi-Fi 8.8.8.8 8.8.4.4
func SetDnsServers(ctx context.Context, servers []string) error {
	var str = fmt.Sprintf("networksetup -setdnsservers Wi-Fi empty")
	if len(servers) != 0 {
		str = fmt.Sprintf("networksetup -setdnsservers Wi-Fi %s", strings.Join(servers, " "))
	}
	log.Debug(str)
	split := strings.Split(str, " ")
	output, err := exec.CommandContext(ctx, split[0], split[1:]...).Output()
	if err != nil {
		return errors.Wrap(err, string(output))
	}
	return nil
}

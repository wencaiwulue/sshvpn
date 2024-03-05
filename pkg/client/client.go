package client

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/miekg/dns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	pkgtun "github.com/wencaiwulue/kubevpn/v2/pkg/tun"
	pkgutil "github.com/wencaiwulue/kubevpn/v2/pkg/util"

	"github.com/wencaiwulue/tlstunnel/pkg/config"
	pkgdns "github.com/wencaiwulue/tlstunnel/pkg/dns"
	"github.com/wencaiwulue/tlstunnel/pkg/tun"
)

func Connect(ctx context.Context, CIDRs []string, conf pkgutil.SshConfig, mode config.ProxyType, stackType config.StackType) error {
	resolvConf, err := getDnsConfig(conf)
	if err != nil {
		return err
	}

	var routes []types.Route
	for _, r := range CIDRs {
		_, c, err := net.ParseCIDR(r)
		if err != nil {
			return fmt.Errorf("invalid cidr %s, err: %v", r, err)
		}
		routes = append(routes, types.Route{Dst: *c})
	}
	for _, server := range resolvConf.Servers {
		routes = append(routes, types.Route{
			Dst: net.IPNet{
				IP:   net.ParseIP(server),
				Mask: net.CIDRMask(len(net.ParseIP(server))*8, len(net.ParseIP(server))*8),
			},
		})
	}
	ipv4 := net.ParseIP("223.253.0.1")
	ipv6 := net.ParseIP("efff:ffff:ffff:ffff:ffff:ffff:ffff:8888")
	addr4 := (&net.IPNet{IP: ipv4, Mask: net.CIDRMask(32, 32)}).String()
	addr6 := (&net.IPNet{IP: ipv6, Mask: net.CIDRMask(128, 128)}).String()
	tunConf := pkgtun.Config{
		MTU:    1500,
		Routes: routes,
	}
	switch stackType {
	case config.SingleStackIPv4:
		tunConf.Addr = addr4
	case config.SingleStackIPv6:
		tunConf.Addr6 = addr6
	case config.DualStack:
		tunConf.Addr = addr4
		tunConf.Addr6 = addr6
	}
	listener, err := pkgtun.Listener(tunConf)
	if err != nil {
		return err
	}
	defer listener.Close()
	tunConn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer tunConn.Close()
	addr := listener.Addr().(*net.IPAddr)
	device, err := pkgutil.GetTunDevice(addr.IP)
	if err != nil {
		return err
	}

	err = pkgdns.Append(ctx, *resolvConf, device)
	if err != nil {
		return err
	}
	defer func() {
		err := pkgdns.Remove(context.Background(), *resolvConf, device)
		if err != nil {
			log.Error(err)
		}
	}()
	tcpAddr, udpAddr, err := getForwardAddr(ctx, conf)
	if err != nil {
		return err
	}
	endpoint := tun.NewTunEndpoint(ctx, tunConn, uint32(tunConf.MTU))
	stack := NewStack(ctx, endpoint, device, tcpAddr, udpAddr)
	go stack.Wait()

	log.Infof("you can use VPN now~")
	<-ctx.Done()
	return ctx.Err()
}

func getForwardAddr(ctx context.Context, conf pkgutil.SshConfig) (string, string, error) {
	tcpPort, err := pkgutil.GetAvailableTCPPortOrDie()
	if err != nil {
		return "", "", err
	}
	udpPort, err := pkgutil.GetAvailableUDPPortOrDie()
	if err != nil {
		return "", "", err
	}

	portPair := []string{
		fmt.Sprintf("%d:%d", tcpPort, config.TCPPort),
		fmt.Sprintf("%d:%d", udpPort, config.UDPPort),
	}
	err = portMap(ctx, &conf, portPair)
	if err != nil {
		return "", "", err
	}
	tcpAddr := fmt.Sprintf("tcp://127.0.0.1:%d", tcpPort)
	udpAddr := fmt.Sprintf("tcp://127.0.0.1:%d", udpPort)
	return tcpAddr, udpAddr, nil
}

func getDnsConfig(conf pkgutil.SshConfig) (*dns.ClientConfig, error) {
	client, err := pkgutil.DialSshRemote(&conf)
	if err != nil {
		return nil, err
	}
	stdout, errout, err := pkgutil.RemoteRun(client, "cat /etc/resolv.conf", nil)
	if err != nil {
		return nil, errors.Wrap(err, string(errout))
	}
	resolvConf, err := dns.ClientConfigFromReader(bytes.NewBufferString(string(stdout)))
	if err != nil {
		return nil, errors.Wrap(err, string(stdout))
	}
	return resolvConf, nil
}

// portPair is local:remote
func portMap(ctx context.Context, conf *pkgutil.SshConfig, portPair []string) error {
	for _, s := range portPair {
		ports := strings.Split(s, ":")
		if len(ports) != 2 {
			return fmt.Errorf("port pair %s is invalid", s)
		}
		local, err := netip.ParseAddrPort(net.JoinHostPort("127.0.0.1", ports[0]))
		if err != nil {
			return err
		}
		var remote netip.AddrPort
		remote, err = netip.ParseAddrPort(net.JoinHostPort("127.0.0.1", ports[1]))
		if err != nil {
			return err
		}
		err = pkgutil.PortMapUntil(ctx, conf, remote, local)
		if err != nil {
			return err
		}
	}
	return nil
}

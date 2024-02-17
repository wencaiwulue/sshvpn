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
	"golang.org/x/crypto/ssh"

	"github.com/wencaiwulue/tlstunnel/pkg/config"
	pkgdns "github.com/wencaiwulue/tlstunnel/pkg/dns"
	"github.com/wencaiwulue/tlstunnel/pkg/tun"
)

func Connect(ctx context.Context, CIDRs []string, conf pkgutil.SshConfig) error {
	tcpPort, err := pkgutil.GetAvailableTCPPortOrDie()
	if err != nil {
		return err
	}
	udpPort, err := pkgutil.GetAvailableUDPPortOrDie()
	if err != nil {
		return err
	}

	portPair := []string{
		fmt.Sprintf("%d:%d", tcpPort, config.TCPPort),
		fmt.Sprintf("%d:%d", udpPort, config.UDPPort),
	}
	client, err := pkgutil.DialSshRemote(ctx, &conf)
	if err != nil {
		return err
	}
	err = portMap(ctx, client, portPair)
	if err != nil {
		return err
	}
	output, out, err := pkgutil.RemoteRun(client, "cat /etc/resolv.conf", nil)
	if err != nil {
		return errors.Wrap(err, string(out))
	}
	resolvConf, err := dns.ClientConfigFromReader(bytes.NewBufferString(string(output)))
	if err != nil {
		return err
	}
	err = pkgdns.Append(ctx, resolvConf.Servers)
	if err != nil {
		return err
	}
	defer func() {
		err := pkgdns.Remove(context.Background(), resolvConf.Servers)
		if err != nil {
			log.Error(err)
		}
	}()

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
	routes = append(routes,
		types.Route{
			Dst: net.IPNet{
				IP:   net.ParseIP("142.250.0.0"),
				Mask: net.CIDRMask(16, 32),
			},
		},
	)
	ipv4 := net.IPv4(223, 253, 0, 1)
	ipv6 := net.IP{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 1}
	tunConf := pkgtun.Config{
		Addr:   (&net.IPNet{IP: ipv4, Mask: net.CIDRMask(32, 32)}).String(),
		Addr6:  (&net.IPNet{IP: ipv6, Mask: net.CIDRMask(128, 128)}).String(),
		MTU:    1350,
		Routes: routes,
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

	tcpAddr := fmt.Sprintf("tcp://127.0.0.1:%d", tcpPort)
	udpAddr := fmt.Sprintf("tcp://127.0.0.1:%d", udpPort)
	endpoint := tun.NewTunEndpoint(ctx, tunConn, uint32(tunConf.MTU))
	stack := NewStack(ctx, endpoint, device, tcpAddr, udpAddr)
	go stack.Wait()

	log.Infof("you can use VPN now~")
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

// portPair is local:remote
func portMap(ctx context.Context, client *ssh.Client, portPair []string) error {
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
		err = pkgutil.PortMapUntil(ctx, client, remote, local)
		if err != nil {
			return err
		}
	}
	return nil
}

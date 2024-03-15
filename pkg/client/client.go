package client

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"time"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"github.com/libp2p/go-netroute"
	"github.com/miekg/dns"
	miekgdns "github.com/miekg/dns"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	pkgtun "github.com/wencaiwulue/kubevpn/v2/pkg/tun"
	pkgutil "github.com/wencaiwulue/kubevpn/v2/pkg/util"

	"github.com/wencaiwulue/tlstunnel/pkg/config"
	pkgdns "github.com/wencaiwulue/tlstunnel/pkg/dns"
	"github.com/wencaiwulue/tlstunnel/pkg/tun"
	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

func Connect(ctx context.Context, CIDRs []types.Route, conf pkgutil.SshConfig, mode config.ProxyMode, stackType config.StackType) error {
	remoteDnsConfig, err := getRemoteDnsConfig(conf)
	if err != nil {
		return err
	}
	for _, server := range remoteDnsConfig.Servers {
		CIDRs = append(CIDRs, types.Route{Dst: net.IPNet{IP: net.ParseIP(server), Mask: util.GetMask(net.ParseIP(server))}})
	}
	remoteServer := util.GetServer(*remoteDnsConfig)
	tunConf := pkgtun.Config{
		MTU:    1500,
		Routes: CIDRs,
	}
	switch stackType {
	case config.SingleStackIPv4:
		tunConf.Addr = config.Addr4
	case config.SingleStackIPv6:
		tunConf.Addr6 = config.Addr6
	case config.DualStack:
		tunConf.Addr = config.Addr4
		tunConf.Addr6 = config.Addr6
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
	servers, err := pkgdns.GetDnsServers(ctx, device)
	if err != nil {
		return err
	}
	localServer := util.GetServer(*servers)
	if mode == config.ProxyModeSplit {
		remoteDnsConfig.Servers = []string{"127.0.0.1"}
		remoteDnsConfig.Port = strconv.Itoa(53)
	}
	err = pkgdns.Append(ctx, *remoteDnsConfig, device)
	if err != nil {
		return err
	}
	defer func() {
		err := pkgdns.Remove(context.Background(), *remoteDnsConfig, device)
		if err != nil {
			log.Error(err)
		}
	}()
	tcpAddr, udpAddr, err := getForwardAddr(ctx, conf)
	if err != nil {
		return err
	}
	endpoint := tun.NewTunEndpoint(ctx, tunConn, uint32(tunConf.MTU))
	r, err := netroute.New()
	if err != nil {
		log.Fatal(err)
	}
	var peekPacket = func(packet []byte) {
		peekPacket(packet, r, mode, device.Name)
	}
	var peekDns = func(dns *miekgdns.Msg) {
		peekDns(device.Name, dns)
	}
	err = setupDnsServer(ctx, mode, peekDns, remoteServer, localServer)
	if err != nil {
		return err
	}
	stack := NewStack(ctx, endpoint, tcpAddr, udpAddr, peekPacket)
	go stack.Wait()

	log.Infof("you can use VPN now~")
	<-ctx.Done()
	return ctx.Err()
}

func setupDnsServer(ctx context.Context, mode config.ProxyMode, peekDns func(dns *miekgdns.Msg), remoteServer, localServer string) error {
	// www.youtube.com. --> youtube.com
	var needForward = func(domain string) bool {
		domain = strings.TrimPrefix(strings.TrimSuffix(domain, "."), "www.")
		if mode == config.ProxyModeSplit && util.Set.Has(domain) {
			return true
		}
		return false
	}
	address := "localhost:53"
	for _, network := range []string{"tcp", "udp"} {
		go func(network string) {
			for ctx.Err() == nil {
				_ = pkgdns.NewDNSServer(network, address, needForward, peekDns, remoteServer, localServer)
				time.Sleep(time.Second * 5)
			}
		}(network)
	}
	return nil
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

func getRemoteDnsConfig(conf pkgutil.SshConfig) (*dns.ClientConfig, error) {
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

func peekPacket(data []byte, r routing.Router, mode config.ProxyMode, tunName string) {
	packet := gopacket.NewPacket(data, layers.LayerTypeDNS, gopacket.Default)
	dnsPacket, ok := packet.ApplicationLayer().(*layers.DNS)
	if !ok {
		return
	}
	for _, answer := range dnsPacket.Answers {
		if answer.IP == nil || len(answer.Name) == 0 {
			continue
		}
		if mode == config.ProxyModeSplit && !util.Set.Has(string(answer.Name)) {
			continue
		}
		// if route is right, not need add route
		ife, _, _, errs := r.Route(answer.IP)
		if errs == nil && tunName == ife.Name {
			continue
		}
		addRoute(tunName, answer.IP, string(answer.Name))
	}
}

func peekDns(tunName string, dns *miekgdns.Msg) {
	if dns == nil {
		return
	}
	for _, rr := range dns.Answer {
		switch a := rr.(type) {
		case *miekgdns.A:
			if ip := a.A; ip != nil && !ip.IsLoopback() {
				addRoute(tunName, ip, a.Hdr.Name)
			}
		case *miekgdns.AAAA:
			if ip := a.AAAA; ip != nil && !ip.IsLoopback() {
				addRoute(tunName, ip, a.Hdr.Name)
			}
		}
	}
}

func addRoute(tunName string, ip net.IP, domain string) {
	log.Debugf("Name: %s --> IP: %s", domain, ip.String())
	err := pkgtun.AddRoutes(tunName, types.Route{
		Dst: net.IPNet{
			IP:   ip,
			Mask: util.GetMask(ip),
		},
		GW: nil,
	})
	if err != nil {
		log.Warnf("failed to add to route: %v", err)
	}
}

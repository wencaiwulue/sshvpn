package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/netip"
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
	//remoteServer := util.GetServer(*remoteDnsConfig)
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
	localServer := util.GetServer(*servers, remoteDnsConfig)
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
	var peekRequest = func(packet []byte, writer io.Writer) bool {
		return peekRequest(packet, mode, localServer, writer)
	}
	//var peekDns = func(dns *miekgdns.Msg) {
	//	peekDns(device.Name, dns)
	//}
	//err = setupDnsServer(ctx, mode, peekDns, remoteServer, localServer)
	//if err != nil {
	//	return err
	//}
	stack := NewStack(ctx, endpoint, tcpAddr, udpAddr, peekRequest, peekPacket)
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

func peekRequest(data []byte, mode config.ProxyMode, localServer string, writer io.Writer) (handled bool) {
	if mode != config.ProxyModeSplit {
		return
	}
	packet := gopacket.NewPacket(data, layers.LayerTypeDNS, gopacket.Default)
	dnsPacket, ok := packet.ApplicationLayer().(*layers.DNS)
	if !ok {
		return
	}
	name := string(dnsPacket.Questions[0].Name)
	name = strings.TrimSuffix(name, ".")
	name = strings.TrimPrefix(name, "www.")
	if util.Set.Has(name) {
		return
	}

	var q []miekgdns.Question
	for _, question := range dnsPacket.Questions {
		q = append(q, miekgdns.Question{
			Name:   string(question.Name) + ".",
			Qtype:  uint16(question.Type),
			Qclass: uint16(question.Class),
		})
	}
	msg := miekgdns.Msg{
		MsgHdr: miekgdns.MsgHdr{
			Id:     dnsPacket.ID,
			Opcode: int(dnsPacket.OpCode),
			//Authoritative:      dnsPacket.Authorities,
			//Truncated:          dnsPacket.,
			RecursionDesired:   false,
			RecursionAvailable: false,
			Zero:               false,
			AuthenticatedData:  false,
			CheckingDisabled:   false,
			Rcode:              int(dnsPacket.ResponseCode),
		},
		Question: q,
		Answer:   nil,
		Ns:       nil,
		Extra:    nil,
	}
	client := miekgdns.Client{Net: "udp", SingleInflight: true, Timeout: time.Second * 30}
	answer, _, err := client.ExchangeContext(context.Background(), &msg, localServer)
	if err != nil {
		return
	}
	for _, rr := range answer.Answer {
		switch a := rr.(type) {
		case *miekgdns.A:
			dnsPacket.Answers = append(dnsPacket.Answers,
				layers.DNSResourceRecord{
					Name:  []byte(a.Hdr.Name),
					Type:  layers.DNSType(a.Hdr.Rrtype),
					Class: layers.DNSClass(a.Hdr.Class),
					Data:  a.A,
				})
		case *miekgdns.AAAA:
			dnsPacket.Answers = append(dnsPacket.Answers,
				layers.DNSResourceRecord{
					Name:  []byte(a.Hdr.Name),
					Type:  layers.DNSType(a.Hdr.Rrtype),
					Class: layers.DNSClass(a.Hdr.Class),
					Data:  a.AAAA,
				})

		case *miekgdns.PTR:
			dnsPacket.Answers = append(dnsPacket.Answers,
				layers.DNSResourceRecord{
					Name:  []byte(a.Hdr.Name),
					Type:  layers.DNSType(a.Hdr.Rrtype),
					Class: layers.DNSClass(a.Hdr.Class),
					Data:  []byte(a.Ptr),
				})
		case *miekgdns.NSAPPTR:
			dnsPacket.Answers = append(dnsPacket.Answers,
				layers.DNSResourceRecord{
					Name:  []byte(a.Hdr.Name),
					Type:  layers.DNSType(a.Hdr.Rrtype),
					Class: layers.DNSClass(a.Hdr.Class),
					Data:  []byte(a.Ptr),
				})
		case *miekgdns.CNAME:
			dnsPacket.Answers = append(dnsPacket.Answers,
				layers.DNSResourceRecord{
					Name:  []byte(a.Hdr.Name),
					Type:  layers.DNSType(a.Hdr.Rrtype),
					Class: layers.DNSClass(a.Hdr.Class),
					Data:  []byte(a.Target),
				})
		}
	}
	for _, rr := range answer.Ns {
		switch a := rr.(type) {
		case *miekgdns.SOA:
			dnsPacket.Answers = append(dnsPacket.Answers,
				layers.DNSResourceRecord{
					Name:  []byte(a.Hdr.Name),
					Type:  layers.DNSType(a.Hdr.Rrtype),
					Class: layers.DNSClass(a.Hdr.Class),
					Data:  []byte(a.Ns),
				})
		}
	}
	buf := gopacket.NewSerializeBuffer()
	opts := gopacket.SerializeOptions{FixLengths: true}
	err = gopacket.SerializeLayers(buf, opts, dnsPacket)
	if err != nil {
		return
	}
	i := buf.Bytes()
	_, err = writer.Write(i)
	return true
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

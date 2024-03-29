package client

import (
	"context"
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/packetsocket"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv6"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/icmp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/raw"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
)

var _ stack.UniqueID = (*id)(nil)

type id struct {
}

func (i id) UniqueID() uint64 {
	return 1
}

func NewStack(ctx context.Context, tun stack.LinkEndpoint, device *net.Interface, tcpAddr, udpAddr string) *stack.Stack {
	s := stack.New(stack.Options{
		NetworkProtocols: []stack.NetworkProtocolFactory{
			ipv4.NewProtocol,
			ipv6.NewProtocol,
		},
		TransportProtocols: []stack.TransportProtocolFactory{
			tcp.NewProtocol,
			udp.NewProtocol,
			icmp.NewProtocol4,
			icmp.NewProtocol6,
		},
		Clock:                    tcpip.NewStdClock(),
		AllowPacketEndpointWrite: true,
		HandleLocal:              false,
		// Enable raw sockets for users with sufficient
		// privileges.
		RawFactory: raw.EndpointFactory{},
		UniqueID:   id{},
	})
	// set handler for TCP UDP ICMP
	s.SetTransportProtocolHandler(tcp.ProtocolNumber, TCPHandler(s, tcpAddr))
	s.SetTransportProtocolHandler(udp.ProtocolNumber, UDPHandler(s, device, udpAddr))
	//s.SetTransportProtocolHandler(icmp.ProtocolNumber4, handler.ICMPHandler(s))
	//s.SetTransportProtocolHandler(icmp.ProtocolNumber6, handler.ICMP6Handler(s))

	s.SetRouteTable([]tcpip.Route{
		{
			Destination: header.IPv4EmptySubnet,
			NIC:         1,
		},
		{
			Destination: header.IPv6EmptySubnet,
			NIC:         1,
		},
	})

	s.CreateNICWithOptions(1, packetsocket.New(tun), stack.NICOptions{
		Disabled: false,
		Context:  ctx,
	})
	s.SetPromiscuousMode(1, true)
	s.SetSpoofing(1, true)

	// Enable SACK Recovery.
	{
		opt := tcpip.TCPSACKEnabled(true)
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &opt); err != nil {
			log.Fatal(fmt.Errorf("SetTransportProtocolOption(%d, &%T(%t)): %s", tcp.ProtocolNumber, opt, opt, err))
		}
	}

	// Set default TTLs as required by socket/netstack.
	{
		opt := tcpip.DefaultTTLOption(64)
		if err := s.SetNetworkProtocolOption(ipv4.ProtocolNumber, &opt); err != nil {
			log.Fatal(fmt.Errorf("SetNetworkProtocolOption(%d, &%T(%d)): %s", ipv4.ProtocolNumber, opt, opt, err))
		}
		if err := s.SetNetworkProtocolOption(ipv6.ProtocolNumber, &opt); err != nil {
			log.Fatal(fmt.Errorf("SetNetworkProtocolOption(%d, &%T(%d)): %s", ipv6.ProtocolNumber, opt, opt, err))
		}
	}

	// Enable Receive Buffer Auto-Tuning.
	{
		opt := tcpip.TCPModerateReceiveBufferOption(true)
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &opt); err != nil {
			log.Fatal(fmt.Errorf("SetTransportProtocolOption(%d, &%T(%t)): %s", tcp.ProtocolNumber, opt, opt, err))
		}
	}

	{
		if err := s.SetForwardingDefaultAndAllNICs(ipv4.ProtocolNumber, true); err != nil {
			log.Fatal(fmt.Errorf("set ipv4 forwarding: %s", err))
		}
		if err := s.SetForwardingDefaultAndAllNICs(ipv6.ProtocolNumber, true); err != nil {
			log.Fatal(fmt.Errorf("set ipv6 forwarding: %s", err))
		}
	}

	{
		option := tcpip.TCPModerateReceiveBufferOption(true)
		if err := s.SetTransportProtocolOption(tcp.ProtocolNumber, &option); err != nil {
			log.Fatal(fmt.Errorf("set TCP moderate receive buffer: %s", err))
		}
	}
	return s
}

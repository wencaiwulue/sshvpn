package client

import (
	"context"
	"io"
	"net"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/routing"
	"github.com/libp2p/go-netroute"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/kubevpn/v2/pkg/config"
	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
	"github.com/wencaiwulue/kubevpn/v2/pkg/tun"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func UDPHandler(s *stack.Stack, device *net.Interface, udpAddr string) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	node, err := core.ParseNode(udpAddr)
	if err != nil {
		log.Debugf("[TUN-UDP] Error: parse gviosr udp forward addr %s: %v", udpAddr, err)
		log.Fatal(err)
	}
	node.Client = &core.Client{
		Connector:   core.GvisorUDPOverTCPTunnelConnector(stack.TransportEndpointID{}),
		Transporter: core.TCPTransporter(),
	}
	forwardChain := core.NewChain(5, node)

	var r routing.Router
	r, err = netroute.New()
	if err != nil {
		log.Fatal(err)
	}

	return udp.NewForwarder(s, func(request *udp.ForwarderRequest) {
		endpointID := request.ID()
		log.Debugf("[TUN-UDP] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
			endpointID.LocalPort, endpointID.LocalAddress.String(), endpointID.RemotePort, endpointID.RemoteAddress.String(),
		)
		w := &waiter.Queue{}
		endpoint, tErr := request.CreateEndpoint(w)
		if tErr != nil {
			log.Debugf("[TUN-UDP] Error: can not create endpoint: %v", tErr)
			return
		}

		ctx := context.Background()
		c, err := forwardChain.Node().Client.Dial(context.Background(), forwardChain.Node().Addr)
		if err != nil {
			log.Debugf("[TUN-TCP] Error: failed to dial remote conn: %v", err)
			return
		}
		if err = core.WriteProxyInfo(c, endpointID); err != nil {
			log.Debugf("[TUN-UDP] Error: can not write proxy info: %v", err)
			return
		}
		remote, err := node.Client.ConnectContext(ctx, c)
		if err != nil {
			log.Debugf("[TUN-UDP] Error: can not connect: %v", err)
			return
		}
		conn := gonet.NewUDPConn(s, w, endpoint)

		go func() {
			defer conn.Close()
			defer remote.Close()
			errChan := make(chan error, 2)
			go func() {
				i := config.LPool.Get().([]byte)[:]
				defer config.LPool.Put(i[:])
				written, err2 := io.CopyBuffer(remote, conn, i)
				log.Debugf("[TUN-UDP] Debug: write length %d data to remote", written)
				errChan <- err2
			}()
			go func() {
				i := config.LPool.Get().([]byte)[:]
				defer config.LPool.Put(i[:])
				var written int
				for {
					n, err3 := remote.Read(i[:])
					if err3 != nil {
						errChan <- err3
						break
					}
					written += n
					addRoute(i, n, r, device.Name)
					_, err3 = conn.Write(i[:n])
					if err3 != nil {
						errChan <- err3
						break
					}
				}
				log.Debugf("[TUN-UDP] Debug: read length %d data from remote", written)
			}()
			err = <-errChan
			if err != nil && !errors.Is(err, io.EOF) {
				log.Debugf("[TUN-UDP] Error: dsiconnect: %s >-<: %s: %v", conn.LocalAddr(), remote.RemoteAddr(), err)
			}
		}()
	}).HandlePacket
}

func addRoute(i []byte, n int, r routing.Router, tunName string) {
	packet := gopacket.NewPacket(i[:n], layers.LayerTypeDNS, gopacket.Default)
	dns, ok := packet.ApplicationLayer().(*layers.DNS)
	if !ok {
		return
	}
	for _, answer := range dns.Answers {
		log.Debugf("Name: %s --> IP: %s", answer.Name, answer.IP.String())
		if answer.IP == nil {
			continue
		}
		// if route is right, not need add route
		iface, _, _, errs := r.Route(answer.IP)
		if errs == nil && tunName == iface.Name {
			continue
		}
		err := tun.AddRoutes(tunName, types.Route{
			Dst: net.IPNet{
				IP:   answer.IP,
				Mask: net.CIDRMask(len(answer.IP)*8, len(answer.IP)*8),
			},
			GW: nil,
		})
		if err != nil {
			log.Warnf("failed to add to route: %v", err)
		}
	}
}

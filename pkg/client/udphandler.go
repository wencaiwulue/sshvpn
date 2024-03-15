package client

import (
	"context"
	"io"
	"net"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/kubevpn/v2/pkg/config"
	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
)

func UDPHandler(s *stack.Stack, udpAddr string, peekPacket func(packet []byte)) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
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
		node, err := core.ParseNode(udpAddr)
		if err != nil {
			log.Debugf("[TUN-UDP] Error: parse gviosr udp forward addr %s: %v", udpAddr, err)
			log.Fatal(err)
		}
		node.Client = &core.Client{
			Connector:   core.GvisorUDPOverTCPTunnelConnector(),
			Transporter: core.TCPTransporter(),
		}
		forwardChain := core.NewChain(5, node)
		var c net.Conn
		c, err = forwardChain.Node().Client.Dial(ctx, forwardChain.Node().Addr)
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
				written, err := io.CopyBuffer(remote, conn, i)
				log.Debugf("[TUN-UDP] Debug: write length %d data to remote", written)
				errChan <- err
			}()
			go func() {
				i := config.LPool.Get().([]byte)[:]
				defer config.LPool.Put(i[:])
				var written int
				for {
					n, err := remote.Read(i[:])
					if err != nil {
						errChan <- err
						break
					}
					written += n
					peekPacket(i[:n])
					_, err = conn.Write(i[:n])
					if err != nil {
						errChan <- err
						break
					}
				}
				log.Debugf("[TUN-UDP] Debug: read length %d data from remote", written)
			}()
			err = <-errChan
			if err != nil && !errors.Is(err, io.EOF) {
				log.Debugf("[TUN-UDP] Error: disconnect: %s >-<: %s: %v", conn.LocalAddr(), remote.RemoteAddr(), err)
			}
		}()
	}).HandlePacket
}

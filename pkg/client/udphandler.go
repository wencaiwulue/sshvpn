package client

import (
	"context"
	"crypto/tls"
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/kubevpn/v2/pkg/config"
	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
)

func UDPHandler(s *stack.Stack, remote string) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	return udp.NewForwarder(s, func(request *udp.ForwarderRequest) {
		w := &waiter.Queue{}
		dial, err := tls.Dial("tcp", remote, ssl.TlsConfigClient)
		if err != nil {
			log.Warningln(err)
			return
		}
		defer dial.Close()
		log.Debugf("[TUN-TCP-CLIENT] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
			request.ID().LocalPort, request.ID().LocalAddress.String(), request.ID().RemotePort, request.ID().RemoteAddress.String(),
		)
		endpoint, t := request.CreateEndpoint(w)
		if t != nil {
			log.Warningln(t)
			return
		}
		defer endpoint.Close()
		conn := gonet.NewUDPConn(s, w, endpoint)
		if err = core.WriteProxyInfo(conn, request.ID()); err != nil {
			log.Warningln(err)
			return
		}
		defer conn.Close()

		connectContext, err := core.GvisorUDPOverTCPTunnelConnector(request.ID()).ConnectContext(context.Background(), dial)
		if err != nil {
			log.Debug(err)
		}

		errChan := make(chan error, 2)
		go func() {
			i := config.LPool.Get().([]byte)[:]
			defer config.LPool.Put(i[:])
			written, err2 := io.CopyBuffer(connectContext, conn, i)
			log.Debugf("[TUN-UDP] Debug: write length %d data to remote", written)
			errChan <- err2
		}()
		go func() {
			i := config.LPool.Get().([]byte)[:]
			defer config.LPool.Put(i[:])
			written, err2 := io.CopyBuffer(conn, connectContext, i)
			log.Debugf("[TUN-UDP] Debug: read length %d data from remote", written)
			errChan <- err2
		}()
		err = <-errChan
		if err != nil && !errors.Is(err, io.EOF) {
			log.Debugf("[TUN-UDP] Error: dsiconnect: %s >-<: %s: %v", conn.LocalAddr(), dial.RemoteAddr(), err)
		}
	}).HandlePacket
}

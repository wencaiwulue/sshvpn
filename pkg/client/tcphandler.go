package client

import (
	"crypto/tls"
	"errors"
	"io"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/kubevpn/v2/pkg/config"
	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
)

func TCPHandler(s *stack.Stack, remote string) func(stack.TransportEndpointID, *stack.PacketBuffer) bool {
	return tcp.NewForwarder(s, 0, 100000, func(request *tcp.ForwarderRequest) {
		defer request.Complete(false)
		log.Debugf("[TUN-TCP-CLIENT] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
			request.ID().LocalPort, request.ID().LocalAddress.String(), request.ID().RemotePort, request.ID().RemoteAddress.String(),
		)
		dial, err := tls.Dial("tcp", remote, ssl.TlsConfigClient)
		if err != nil {
			log.Warningln(err)
			return
		}
		defer dial.Close()
		if err = core.WriteProxyInfo(dial, request.ID()); err != nil {
			log.Warningln(err)
			return
		}

		w := &waiter.Queue{}
		endpoint, t := request.CreateEndpoint(w)
		if t != nil {
			log.Warningln(t)
			return
		}
		defer endpoint.Close()
		conn := gonet.NewTCPConn(w, endpoint)
		defer conn.Close()

		errChan := make(chan error, 2)
		go func() {
			i := config.LPool.Get().([]byte)[:]
			defer config.LPool.Put(i[:])
			written, err := io.CopyBuffer(dial, conn, i)
			log.Debugf("[TUN-TCP] Debug: write length %d data to remote", written)
			errChan <- err
		}()
		go func() {
			i := config.LPool.Get().([]byte)[:]
			defer config.LPool.Put(i[:])
			written, err := io.CopyBuffer(conn, dial, i)
			log.Debugf("[TUN-TCP] Debug: read length %d data from remote", written)
			errChan <- err
		}()
		err = <-errChan
		if err != nil && !errors.Is(err, io.EOF) {
			log.Debugf("[TUN-TCP] Error: dsiconnect: %s >-<: %s: %v", conn.LocalAddr(), dial.RemoteAddr(), err)
		}
	}).HandlePacket
}

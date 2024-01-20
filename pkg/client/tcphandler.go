package client

import (
	"crypto/tls"
	"io"

	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

func TCPHandler(s *stack.Stack, remote string) func(stack.TransportEndpointID, *stack.PacketBuffer) bool {
	return tcp.NewForwarder(s, 0, 100000, func(request *tcp.ForwarderRequest) {
		defer request.Complete(false)
		log.Debugf("[TUN-TCP-CLIENT] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
			request.ID().LocalPort, request.ID().LocalAddress.String(), request.ID().RemotePort, request.ID().RemoteAddress.String(),
		)
		dialer, err := tls.Dial("tcp", remote, ssl.TlsConfigClient)
		if err != nil {
			log.Warningln(err)
			return
		}
		defer dialer.Close()
		if err = util.WriteProxyInfo(dialer, request.ID()); err != nil {
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
		go io.Copy(dialer, conn)
		io.Copy(conn, dialer)
	}).HandlePacket
}

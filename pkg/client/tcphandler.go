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
		log.Infof("[TUN-TCP] Info: %#v", request.ID())
		dialer, err := tls.Dial("tcp", remote, ssl.TlsConfigClient)
		if err != nil {
			log.Warningln(err)
			return
		}
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
		conn := gonet.NewTCPConn(w, endpoint)
		go io.Copy(dialer, conn)
		io.Copy(conn, dialer)
	}).HandlePacket
}

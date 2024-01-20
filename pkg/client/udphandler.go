package client

import (
	"crypto/tls"
	"io"

	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"

	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
	"github.com/wencaiwulue/tlstunnel/pkg/util"
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
		log.Debugf("[TUN-TCP] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
			request.ID().LocalPort, request.ID().LocalAddress.String(), request.ID().RemotePort, request.ID().RemoteAddress.String(),
		)
		endpoint, t := request.CreateEndpoint(w)
		if t != nil {
			log.Warningln(t)
			return
		}
		defer endpoint.Close()
		conn := gonet.NewUDPConn(s, w, endpoint)
		if err = util.WriteProxyInfo(conn, request.ID()); err != nil {
			log.Warningln(err)
			return
		}
		defer conn.Close()
		go io.Copy(conn, dial)
		io.Copy(dial, conn)
	}).HandlePacket
}

package handler

import (
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/udp"
	"gvisor.dev/gvisor/pkg/waiter"
	"io"
	"net"
)

func UDPHandler(s *stack.Stack) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	return udp.NewForwarder(s, func(request *udp.ForwarderRequest) {
		w := &waiter.Queue{}
		dial, err := net.Dial("udp", "")
		if err != nil {
			log.Warningln(err)
			return
		}
		endpoint, t := request.CreateEndpoint(w)
		if t != nil {
			log.Warningln(t)
			return
		}
		conn := gonet.NewUDPConn(s, w, endpoint)
		if err = WriteProxyInfo(conn, request.ID()); err != nil {
			log.Warningln(err)
			return
		}
		go io.Copy(conn, dial)
		io.Copy(dial, conn)
	}).HandlePacket
}

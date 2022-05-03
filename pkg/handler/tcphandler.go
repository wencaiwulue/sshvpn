package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
	"gvisor.dev/gvisor/pkg/tcpip/adapters/gonet"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
	"gvisor.dev/gvisor/pkg/waiter"
	"io"
	"net"
)

func TCPHandler(s *stack.Stack) func(stack.TransportEndpointID, *stack.PacketBuffer) bool {
	return tcp.NewForwarder(s, 0, 100000, func(request *tcp.ForwarderRequest) {
		defer request.Complete(false)
		log.Infof("[TUN-TCP] Info: %#v", request.ID())
		dialer, err := tls.Dial("tcp", "localhost:10800", ssl.TlsConfigClient)
		if err != nil {
			log.Warningln(err)
			return
		}
		if err = WriteProxyInfo(dialer, request.ID()); err != nil {
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

func WriteProxyInfo(conn net.Conn, id stack.TransportEndpointID) error {
	var b bytes.Buffer
	i := make([]byte, 2)
	binary.BigEndian.PutUint16(i, id.LocalPort)
	b.Write(i)
	binary.BigEndian.PutUint16(i, id.RemotePort)
	b.Write(i)
	local := string(id.LocalAddress)
	b.WriteByte(byte(len(local)))
	b.WriteString(local)
	remote := string(id.RemoteAddress)
	b.WriteByte(byte(len(remote)))
	b.WriteString(remote)
	_, err := b.WriteTo(conn)
	return err
}

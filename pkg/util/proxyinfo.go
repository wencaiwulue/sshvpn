package util

import (
	"bytes"
	"encoding/binary"
	"io"
	"net"

	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

// WriteProxyInfo write proxy info
func WriteProxyInfo(conn net.Conn, id stack.TransportEndpointID) error {
	var b bytes.Buffer
	i := make([]byte, 2)
	binary.BigEndian.PutUint16(i, id.LocalPort)
	b.Write(i)
	binary.BigEndian.PutUint16(i, id.RemotePort)
	b.Write(i)
	local := id.LocalAddress
	b.WriteByte(byte(local.Len()))
	b.Write(local.AsSlice())
	remote := id.RemoteAddress
	b.WriteByte(byte(remote.Len()))
	b.Write(remote.AsSlice())
	_, err := b.WriteTo(conn)
	return err
}

// ParseInfo parse proxy info [20]byte
func ParseInfo(conn net.Conn) (id stack.TransportEndpointID, err error) {
	var n int
	var port = make([]byte, 2)

	// local port
	if n, err = io.ReadFull(conn, port); err != nil || n != 2 {
		return
	}
	id.LocalPort = binary.BigEndian.Uint16(port)

	// remote port
	if n, err = io.ReadFull(conn, port); err != nil || n != 2 {
		return
	}
	id.RemotePort = binary.BigEndian.Uint16(port)

	// local address
	if n, err = io.ReadFull(conn, port[:1]); err != nil || n != 1 {
		return
	}
	var localAddress = make([]byte, port[0])
	if n, err = io.ReadFull(conn, localAddress); err != nil || n != len(localAddress) {
		return
	}
	id.LocalAddress = tcpip.AddrFromSlice(localAddress)

	// remote address
	if n, err = io.ReadFull(conn, port[:1]); err != nil || n != 1 {
		return
	}
	var remoteAddress = make([]byte, port[0])
	if n, err = io.ReadFull(conn, remoteAddress); err != nil || n != len(remoteAddress) {
		return
	}
	id.RemoteAddress = tcpip.AddrFromSlice(remoteAddress)
	return
}

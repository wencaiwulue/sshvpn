package server

import (
	"errors"
	"io"
	"net"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/kubevpn/v2/pkg/config"

	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

func UDPHandler(conn net.Conn) {
	defer conn.Close()
	// 1, get proxy info
	endpoint, err := util.ParseInfo(conn)
	if err != nil {
		log.Warningln(err)
		return
	}
	log.Debugf("[TUN-TCP-SERVER] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
		endpoint.LocalPort, endpoint.LocalAddress.String(), endpoint.RemotePort, endpoint.RemoteAddress.String(),
	)
	//2, dial proxy
	proxy := &net.UDPAddr{IP: net.ParseIP(endpoint.RemoteAddress.String()), Port: int(endpoint.RemotePort)}
	remote, err2 := net.DialUDP("udp", nil, proxy)
	if err2 != nil {
		log.Warningln(err2)
		return
	}
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
		written, err2 := io.CopyBuffer(conn, remote, i)
		log.Debugf("[TUN-UDP] Debug: read length %d data from remote", written)
		errChan <- err2
	}()
	err = <-errChan
	if err != nil && !errors.Is(err, io.EOF) {
		log.Debugf("[TUN-UDP] Error: dsiconnect: %s >-<: %s: %v", conn.LocalAddr(), remote.RemoteAddr(), err)
	}
}

package server

import (
	"errors"
	"io"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/kubevpn/v2/pkg/config"

	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

func TCPHandler(conn net.Conn) {
	defer conn.Close()
	// 1, get proxy endpoint
	endpoint, err := util.ParseInfo(conn)
	if err != nil {
		log.Warningln(err)
		return
	}
	log.Debugf("[TUN-TCP-SERVER] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
		endpoint.LocalPort, endpoint.LocalAddress.String(), endpoint.RemotePort, endpoint.RemoteAddress.String(),
	)
	// 2, dial proxy
	proxy := net.JoinHostPort(endpoint.LocalAddress.String(), strconv.FormatUint(uint64(endpoint.LocalPort), 10))
	var remote net.Conn
	remote, err = net.DialTimeout("tcp", proxy, time.Second*5)
	if err != nil {
		log.Warningln(err)
		return
	}
	defer remote.Close()
	errChan := make(chan error, 2)
	go func() {
		i := config.LPool.Get().([]byte)[:]
		defer config.LPool.Put(i[:])
		written, err := io.CopyBuffer(remote, conn, i)
		log.Debugf("[TUN-TCP] Debug: write length %d data to remote", written)
		errChan <- err
	}()
	go func() {
		i := config.LPool.Get().([]byte)[:]
		defer config.LPool.Put(i[:])
		written, err := io.CopyBuffer(conn, remote, i)
		log.Debugf("[TUN-TCP] Debug: read length %d data from remote", written)
		errChan <- err
	}()
	err = <-errChan
	if err != nil && !errors.Is(err, io.EOF) {
		log.Debugf("[TUN-TCP] Error: dsiconnect: %s >-<: %s: %v", conn.LocalAddr(), remote.RemoteAddr(), err)
	}
}

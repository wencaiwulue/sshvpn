package server

import (
	"io"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

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
	log.Debugf("[TUN-TCP] Debug: LocalPort: %d, LocalAddress: %s, RemotePort: %d, RemoteAddress %s",
		endpoint.LocalPort, endpoint.LocalAddress.String(), endpoint.RemotePort, endpoint.RemoteAddress.String(),
	)
	// 2, dial proxy
	proxy := net.JoinHostPort(endpoint.LocalAddress.String(), strconv.FormatUint(uint64(endpoint.LocalPort), 10))
	dial, err2 := net.DialTimeout("tcp", proxy, time.Second*5)
	if err2 != nil {
		log.Warningln(err2)
		return
	}
	defer dial.Close()
	go io.Copy(conn, dial)
	io.Copy(dial, conn)
}

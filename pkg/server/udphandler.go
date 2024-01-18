package server

import (
	"io"
	"net"

	log "github.com/sirupsen/logrus"

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
	//2, dial proxy
	proxy := &net.UDPAddr{IP: net.ParseIP(endpoint.RemoteAddress.String()), Port: int(endpoint.RemotePort)}
	dial, err2 := net.DialUDP("udp", nil, proxy)
	if err2 != nil {
		log.Warningln(err2)
		return
	}
	go io.Copy(conn, dial)
	io.Copy(dial, conn)
}

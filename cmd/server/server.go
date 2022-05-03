package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
	"github.com/wencaiwulue/tlstunnel/pkg/util"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func main() {
	// setup dns server
	go func() {
		if err := util.NewDNSServer("0.0.0.0:53"); err != nil {
			log.Fatal(err)
		}
	}()

	tcpListener, err := tls.Listen("tcp", "localhost:10800", ssl.TlsConfigServer)
	if err != nil {
		log.Fatal(err)
	}
	//udpListener, err := tls.Listen("udp", "localhost:1080", ssl.TlsConfigServer)
	//if err != nil {
	//	log.Fatal(err)
	//}
	go func() {
		for {
			accept, err := tcpListener.Accept()
			if err != nil {
				log.Warningln(err)
				continue
			}
			go func(conn net.Conn) {
				defer conn.Close()
				// 1, get proxy info
				a, err3 := parseInfo(conn)
				if err3 != nil {
					log.Warningln(err3)
					return
				}
				// 2, dial proxy
				s := net.IP(a.LocalAddress).String()
				itoa := strconv.FormatUint(uint64(a.LocalPort), 10)
				dial, err2 := net.DialTimeout("tcp", net.JoinHostPort(s, itoa), time.Second*5)
				if err2 != nil {
					log.Warningln(err2)
					return
				}
				go io.Copy(conn, dial)
				io.Copy(dial, conn)
			}(accept)
		}
	}()
	select {}

	//go func() {
	//	for {
	//		accept, err := udpListener.Accept()
	//		if err != nil {
	//			log.Warningln(err)
	//			continue
	//		}
	//		go func(conn net.Conn) {
	//			defer conn.Close()
	//			// 1, get proxy info
	//			a, err3 := parseInfo(conn)
	//			if err3 != nil {
	//				log.Warningln(err3)
	//				return
	//			}
	//			// 2, dial proxy
	//			dial, err2 := net.Dial("udp", net.JoinHostPort(string(a.RemoteAddress), strconv.Itoa(int(a.RemotePort))))
	//			if err2 != nil {
	//				log.Warningln(err2)
	//				return
	//			}
	//			go io.Copy(conn, dial)
	//			io.Copy(dial, conn)
	//		}(accept)
	//	}
	//}()
}

// parse proxy info [20]byte
func parseInfo(conn net.Conn) (id stack.TransportEndpointID, err error) {
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
	id.LocalAddress = tcpip.Address(localAddress)

	// remote address
	if n, err = io.ReadFull(conn, port[:1]); err != nil || n != 1 {
		return
	}
	var remoteAddress = make([]byte, port[0])
	if n, err = io.ReadFull(conn, remoteAddress); err != nil || n != len(remoteAddress) {
		return
	}
	id.RemoteAddress = tcpip.Address(remoteAddress)
	return
}

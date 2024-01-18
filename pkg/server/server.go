package server

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/wencaiwulue/tlstunnel/pkg/ssl"
)

func Serve(ctx context.Context, tcpPort int, udpPort int) error {
	errChan := make(chan error, 2)

	// 1) setup tcp server
	go func() {
		err := tcpListener(tcpPort)
		if err != nil {
			errChan <- err
		}
	}()

	// 2) setup udp server
	go func() {
		err := udpListener(udpPort)
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func tcpListener(tcpPort int) error {
	listener, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", tcpPort), ssl.TlsConfigServer)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go TCPHandler(conn)
	}
}

func udpListener(udpPort int) error {
	listener, err := tls.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", udpPort), ssl.TlsConfigServer)
	if err != nil {
		return err
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go UDPHandler(conn)
	}
}

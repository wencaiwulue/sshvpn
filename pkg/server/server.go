package server

import (
	"context"
	"fmt"

	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
)

func Serve(ctx context.Context, tcpPort int, udpPort int) error {
	errChan := make(chan error, 2)

	// 1) setup tcp server
	go func() {
		errChan <- tcpListener(ctx, tcpPort)
	}()

	// 2) setup udp server
	go func() {
		errChan <- udpListener(ctx, udpPort)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-errChan:
		return err
	}
}

func tcpListener(ctx context.Context, tcpPort int) error {
	listener, err := core.GvisorTCPListener(fmt.Sprintf("0.0.0.0:%d", tcpPort))
	if err != nil {
		return err
	}
	defer listener.Close()
	for ctx.Err() == nil {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go TCPHandler(ctx, conn)
	}
	return ctx.Err()
}

func udpListener(ctx context.Context, udpPort int) error {
	listener, err := core.GvisorUDPListener(fmt.Sprintf("0.0.0.0:%d", udpPort))
	if err != nil {
		return err
	}
	defer listener.Close()
	for ctx.Err() == nil {
		conn, err := listener.Accept()
		if err != nil {
			return err
		}
		go UDPHandler(ctx, conn)
	}
	return ctx.Err()
}

package server

import (
	"context"
	"net"

	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
)

func TCPHandler(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	core.GvisorTCPHandler().Handle(ctx, conn)
}

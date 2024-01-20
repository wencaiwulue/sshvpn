package server

import (
	"context"
	"net"

	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
)

func UDPHandler(conn net.Conn) {
	defer conn.Close()
	core.GvisorUDPHandler().Handle(context.Background(), conn)
}

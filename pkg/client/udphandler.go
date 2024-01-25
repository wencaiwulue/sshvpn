package client

import (
	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func UDPHandler(s *stack.Stack, remote string) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	core.GvisorUDPForwardAddr = remote
	return core.UDPForwarder(s)
}

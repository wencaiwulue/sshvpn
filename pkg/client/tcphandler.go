package client

import (
	"github.com/wencaiwulue/kubevpn/v2/pkg/core"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func TCPHandler(s *stack.Stack, remote string) func(stack.TransportEndpointID, *stack.PacketBuffer) bool {
	core.GvisorTCPForwardAddr = remote
	return core.TCPForwarder(s)
}

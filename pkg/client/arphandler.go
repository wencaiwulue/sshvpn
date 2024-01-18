package client

import "gvisor.dev/gvisor/pkg/tcpip/stack"

func ARPHandler(s *stack.Stack) func(stack.TransportEndpointID, *stack.PacketBuffer) bool {
	return func(id stack.TransportEndpointID, buffer *stack.PacketBuffer) bool {
		return true
	}
}

package handler

import (
	log "github.com/sirupsen/logrus"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/header/parse"
	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
)

func ICMP6Handler(s *stack.Stack) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
	return func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
		log.Infof("[ICMP] Receive a icmp package, SRC: %s, DST: %s", id.LocalAddress, id.RemoteAddress)
		if id.LocalAddress.String() == id.RemoteAddress.String() {
			_, view := handleICMP6(pkt)
			s.WriteRawPacket(1, ipv4.ProtocolNumber, view)
		} else {
			// todo forward it to remote
			vv := pkt.Data().ExtractVV()
			_ = vv.ToOwnedView()
		}
		return true
	}
}

func handleICMP6(pkt *stack.PacketBuffer) (*stack.PacketBuffer, buffer.VectorisedView) {
	replyData := stack.PayloadSince(pkt.TransportHeader())
	iph := header.IPv6(pkt.NetworkHeader().View())
	replyHeaderLength := uint8(header.IPv4MinimumSize)
	replyIPHdrBytes := make([]byte, 0, replyHeaderLength)
	replyIPHdrBytes = append(replyIPHdrBytes, iph[:header.IPv4MinimumSize]...)
	replyIPHdrBytes = append(replyIPHdrBytes)
	replyIPHdr := header.IPv6(replyIPHdrBytes)
	replyIPHdr.SetSourceAddress(iph.DestinationAddress())
	replyIPHdr.SetDestinationAddress(iph.SourceAddress())
	replyIPHdr.SetHopLimit(iph.HopLimit())
	replyIPHdr.SetPayloadLength(uint16(len(replyData)))

	replyICMPHdr := header.ICMPv6(replyData)
	replyICMPHdr.SetType(header.ICMPv6EchoReply)
	replyICMPHdr.SetChecksum(0)
	replyICMPHdr.SetChecksum(^header.Checksum(replyData, 0))

	replyVV := buffer.View(replyIPHdr).ToVectorisedView()
	replyVV.AppendView(replyData)
	replyPkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
		ReserveHeaderBytes: header.ICMPv6HeaderSize,
		Data:               replyVV,
	})
	defer replyPkt.DecRef()
	// Populate the network/transport headers in the packet buffer so the
	// ICMP packet goes through IPTables.
	if _, _, _, _, ok := parse.IPv6(replyPkt); !ok {
		panic("expected to parse IPv4 header we just created")
	}
	if ok := parse.ICMPv6(replyPkt); !ok {
		panic("expected to parse ICMPv4 header we just created")
	}
	return replyPkt, replyVV
}

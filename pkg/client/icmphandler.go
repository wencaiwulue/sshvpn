package client

//
//import (
//	log "github.com/sirupsen/logrus"
//	"gvisor.dev/gvisor/pkg/tcpip/buffer"
//	"gvisor.dev/gvisor/pkg/tcpip/header"
//	"gvisor.dev/gvisor/pkg/tcpip/header/parse"
//	"gvisor.dev/gvisor/pkg/tcpip/network/ipv4"
//	"gvisor.dev/gvisor/pkg/tcpip/stack"
//)
//
//func ICMPHandler(s *stack.Stack) func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
//	return func(id stack.TransportEndpointID, pkt *stack.PacketBuffer) bool {
//		log.Infof("[ICMP] Receive a icmp package, SRC: %s, DST: %s", id.LocalAddress, id.RemoteAddress)
//		if id.LocalAddress.String() == id.RemoteAddress.String() {
//			_, view := handleICMP(pkt)
//			s.WriteRawPacket(1, ipv4.ProtocolNumber, view)
//		} else {
//			// todo forward it to remote
//			vv := pkt.Data().ExtractVV()
//			_ = vv.ToOwnedView()
//		}
//		return true
//	}
//}
//
//func handleICMP(pkt *stack.PacketBuffer) (*stack.PacketBuffer, buffer.VectorisedView) {
//	replyData := stack.PayloadSince(pkt.TransportHeader())
//	iph := header.IPv4(pkt.NetworkHeader().View())
//	replyHeaderLength := uint8(header.IPv4MinimumSize)
//	replyIPHdrBytes := make([]byte, 0, replyHeaderLength)
//	replyIPHdrBytes = append(replyIPHdrBytes, iph[:header.IPv4MinimumSize]...)
//	replyIPHdrBytes = append(replyIPHdrBytes)
//	replyIPHdr := header.IPv4(replyIPHdrBytes)
//	replyIPHdr.SetHeaderLength(replyHeaderLength)
//	replyIPHdr.SetSourceAddress(iph.DestinationAddress())
//	replyIPHdr.SetDestinationAddress(iph.SourceAddress())
//	replyIPHdr.SetTTL(iph.TTL())
//	replyIPHdr.SetTotalLength(uint16(len(replyIPHdr) + len(replyData)))
//	replyIPHdr.SetChecksum(0)
//	replyIPHdr.SetChecksum(^replyIPHdr.CalculateChecksum())
//
//	replyICMPHdr := header.ICMPv4(replyData)
//	replyICMPHdr.SetType(header.ICMPv4EchoReply)
//	replyICMPHdr.SetChecksum(0)
//	replyICMPHdr.SetChecksum(^header.Checksum(replyData, 0))
//
//	replyVV := buffer.View(replyIPHdr).ToVectorisedView()
//	replyVV.AppendView(replyData)
//	replyPkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
//		ReserveHeaderBytes: header.IPv4MaximumHeaderSize,
//		Data:               replyVV,
//	})
//	defer replyPkt.DecRef()
//	// Populate the network/transport headers in the packet buffer so the
//	// ICMP packet goes through IPTables.
//	if ok := parse.IPv4(replyPkt); !ok {
//		panic("expected to parse IPv4 header we just created")
//	}
//	if ok := parse.ICMPv4(replyPkt); !ok {
//		panic("expected to parse ICMPv4 header we just created")
//	}
//	return replyPkt, replyVV
//}

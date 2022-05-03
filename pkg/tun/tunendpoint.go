package devtun

import (
	"context"
	"sync"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/tlstunnel/pkg/util"
	"golang.zx2c4.com/wireguard/tun"
	"gvisor.dev/gvisor/pkg/tcpip"
	"gvisor.dev/gvisor/pkg/tcpip/buffer"
	"gvisor.dev/gvisor/pkg/tcpip/header"
	"gvisor.dev/gvisor/pkg/tcpip/link/channel"
	"gvisor.dev/gvisor/pkg/tcpip/stack"
	"gvisor.dev/gvisor/pkg/tcpip/transport/tcp"
)

var _ stack.LinkEndpoint = (*tunEndpoint)(nil)

// tunEndpoint /Users/naison/go/pkg/mod/gvisor.dev/gvisor@v0.0.0-20220422052705-39790bd3a15a/pkg/tcpip/link/tun/device.go:122
type tunEndpoint struct {
	ctx      context.Context
	tun      tun.Device
	once     sync.Once
	endpoint *channel.Endpoint
}

// WritePackets writes packets. Must not be called with an empty list of
// packet buffers.
//
// WritePackets may modify the packet buffers, and takes ownership of the PacketBufferList.
// it is not safe to use the PacketBufferList after a call to WritePackets.
func (e *tunEndpoint) WritePackets(p stack.PacketBufferList) (int, tcpip.Error) {
	return e.endpoint.WritePackets(p)
}

// MTU is the maximum transmission unit for this endpoint. This is
// usually dictated by the backing physical network; when such a
// physical network doesn't exist, the limit is generally 64k, which
// includes the maximum size of an IP packet.
func (e *tunEndpoint) MTU() uint32 {
	mtu, _ := e.tun.MTU()
	return uint32(mtu)
}

// MaxHeaderLength returns the maximum size the data link (and
// lower level layers combined) headers can have. Higher levels use this
// information to reserve space in the front of the packets they're
// building.
func (e *tunEndpoint) MaxHeaderLength() uint16 {
	return 0
}

// LinkAddress returns the link address (typically a MAC) of the
// endpoint.
func (e *tunEndpoint) LinkAddress() tcpip.LinkAddress {
	return e.endpoint.LinkAddress()
}

// Capabilities returns the set of capabilities supported by the
// endpoint.
func (e *tunEndpoint) Capabilities() stack.LinkEndpointCapabilities {
	return e.endpoint.LinkEPCapabilities
}

// Attach attaches the data link layer endpoint to the network-layer
// dispatcher of the stack.
//
// Attach is called with a nil dispatcher when the endpoint's NIC is being
// removed.
func (e *tunEndpoint) Attach(dispatcher stack.NetworkDispatcher) {
	e.endpoint.Attach(dispatcher)
	// queue --> tun
	e.once.Do(func() {
		go func() {
			for {
				func() {
					read := e.endpoint.ReadContext(e.ctx)
					if read != nil {
						size := read.Size()
						views := read.Views()
						views = append([]buffer.View{make(buffer.View, 4)}, views...)

						vView := buffer.NewVectorisedView(size, views)
						_, err := e.tun.Write(vView.ToView(), 4)
						if err != nil {
							log.Warningln(err)
						}
					}
				}()
			}
		}()
		// tun --> dispatcher
		go func() {
			for {
				func() {
					bytes := util.MPool.Get().([]byte)[:]
					defer util.MPool.Put(bytes)
					read, err := e.tun.Read(bytes, 4)
					if err != nil {
						log.Warningln(err)
						return
					}
					if read <= 4 {
						log.Warnf("[TUN]: read from tun length is %d", read)
					}
					// Try to determine network protocol number, default zero.
					var protocol tcpip.NetworkProtocolNumber
					// TUN interface with IFF_NO_PI enabled, thus
					// we need to determine protocol from version field
					version := bytes[4] >> 4
					if version == 4 {
						protocol = header.IPv4ProtocolNumber
					} else if version == 6 {
						protocol = header.IPv6ProtocolNumber
					}

					pkt := stack.NewPacketBuffer(stack.PacketBufferOptions{
						ReserveHeaderBytes: 4,
						Data:               buffer.NewViewFromBytes(bytes[4 : 4+read]).ToVectorisedView(),
					})
					//defer pkt.DecRef()
					e.endpoint.InjectInbound(protocol, pkt)
				}()
			}
		}()
	})
}

// IsAttached returns whether a NetworkDispatcher is attached to the
// endpoint.
func (e *tunEndpoint) IsAttached() bool {
	return e.endpoint.IsAttached()
}

// Wait waits for any worker goroutines owned by the endpoint to stop.
//
// For now, requesting that an endpoint's worker goroutine(s) stop is
// implementation specific.
//
// Wait will not block if the endpoint hasn't started any goroutines
// yet, even if it might later.
func (e *tunEndpoint) Wait() {
	return
}

// ARPHardwareType returns the ARPHRD_TYPE of the link endpoint.
//
// See:
// https://github.com/torvalds/linux/blob/aa0c9086b40c17a7ad94425b3b70dd1fdd7497bf/include/uapi/linux/if_arp.h#L30
func (e *tunEndpoint) ARPHardwareType() header.ARPHardwareType {
	return header.ARPHardwareNone
}

// AddHeader adds a link layer header to the packet if required.
func (e *tunEndpoint) AddHeader(*stack.PacketBuffer) {
	return
}

func NewTunEndpoint(ctx context.Context, tun tun.Device) (stack.LinkEndpoint, error) {
	mtu, err := tun.MTU()
	if err != nil {
		return nil, err
	}
	addr, _ := tcpip.ParseMACAddress("02:03:03:04:05:06")
	return &tunEndpoint{
		ctx:      ctx,
		tun:      tun,
		endpoint: channel.New(tcp.DefaultReceiveBufferSize, uint32(mtu), addr),
	}, nil
}

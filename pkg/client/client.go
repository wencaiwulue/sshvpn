package client

import (
	"context"
	"fmt"
	"net"

	"github.com/containernetworking/cni/pkg/types"
	log "github.com/sirupsen/logrus"
	pkgtun "github.com/wencaiwulue/kubevpn/v2/pkg/tun"
	pkgutil "github.com/wencaiwulue/kubevpn/v2/pkg/util"

	"github.com/wencaiwulue/tlstunnel/pkg/config"
	"github.com/wencaiwulue/tlstunnel/pkg/tun"
	"github.com/wencaiwulue/tlstunnel/pkg/util"
)

func Connect(ctx context.Context, CIDRs []string, conf pkgutil.SshConfig) error {
	tcpPort, err := pkgutil.GetAvailableTCPPortOrDie()
	if err != nil {
		return err
	}
	err = util.Jump(ctx, conf, tcpPort, config.TCPPort)
	if err != nil {
		return err
	}

	udpPort, err := pkgutil.GetAvailableUDPPortOrDie()
	if err != nil {
		return err
	}
	err = util.Jump(ctx, conf, udpPort, config.UDPPort)
	if err != nil {
		return err
	}

	var routes []types.Route
	for _, r := range CIDRs {
		_, c, err := net.ParseCIDR(r)
		if err != nil {
			return fmt.Errorf("invalid cidr %s, err: %v", r, err)
		}
		routes = append(routes, types.Route{Dst: *c})
	}
	routes = append(routes, types.Route{
		Dst: net.IPNet{
			IP:   net.ParseIP("142.250.0.0"),
			Mask: net.CIDRMask(16, 32),
		},
	})
	tunConf := pkgtun.Config{
		Addr:   "223.254.0.1/32",
		Addr6:  "",
		MTU:    1350,
		Routes: routes,
	}
	listener, err := pkgtun.Listener(tunConf)
	if err != nil {
		return err
	}
	defer listener.Close()
	tunConn, err := listener.Accept()
	if err != nil {
		return err
	}
	defer tunConn.Close()
	endpoint := tun.NewTunEndpoint(ctx, tunConn, uint32(tunConf.MTU))

	stack := NewStack(ctx, endpoint, fmt.Sprintf("localhost:%d", tcpPort), fmt.Sprintf("localhost:%d", udpPort))
	go stack.Wait()

	log.Infof("you can use VPN now~")
	select {
	case <-ctx.Done():
		return ctx.Err()
	}
}

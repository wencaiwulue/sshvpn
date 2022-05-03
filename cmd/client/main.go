package main

import (
	"context"
	"flag"
	"net"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/wencaiwulue/tlstunnel/pkg"
	"github.com/wencaiwulue/tlstunnel/pkg/config"
	"github.com/wencaiwulue/tlstunnel/pkg/route"
	devtun "github.com/wencaiwulue/tlstunnel/pkg/tun"
	"golang.zx2c4.com/wireguard/tun"
)

var mode config.ProxyType
var ip string
var pacPath string

func init() {
	flag.StringVar((*string)(&mode), "mode", string(config.ProxyTypeGlobe), "Only support mode Globe or PAC")
	flag.StringVar(&ip, "remote_ip", "", "The ip of remote linux server")
	flag.StringVar(&pacPath, "pac", "", "The path of PAC, can be a url or local path")
	flag.Parse()
}

func main() {
	switch mode {
	case config.ProxyTypeGlobe:
		if len(ip) == 0 {
			log.Fatal("Globe mode, remote ip should not be empty")
		}
	case config.ProxyTypePAC:
		if len(pacPath) == 0 {
			log.Fatal("PAC mode, PAC path should not be empty")
		}
	default:
		log.Fatal("Not support proxy mode " + mode)
	}

	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()
	device, err := tun.CreateTUN("utun", 1350)
	if err != nil {
		log.Fatal(err)
	}
	err = route.AddIPAndRoute(device, net.IPNet{IP: nil, Mask: nil}, nil)
	if err != nil {
		log.Fatal(err)
	}
	endpoint, err := devtun.NewTunEndpoint(ctx, device)
	if err != nil {
		log.Fatal(err)
	}
	pkg.NewStack(ctx, endpoint)
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Kill, os.Interrupt)
	<-signals
	cancelFunc()
}

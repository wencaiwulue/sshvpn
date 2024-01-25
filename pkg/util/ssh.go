package util

import (
	"context"
	"errors"
	"net"
	"net/netip"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/wencaiwulue/kubevpn/v2/pkg/util"
)

func Jump(ctx context.Context, conf util.SshConfig, localPort, remotePort int) error {
	local, err := netip.ParseAddrPort(net.JoinHostPort("127.0.0.1", strconv.Itoa(localPort)))
	if err != nil {
		return err
	}

	remote, err := netip.ParseAddrPort(net.JoinHostPort("127.0.0.1", strconv.Itoa(remotePort)))
	if err != nil {
		return err
	}

	// pre-check network ip connect
	var cli *ssh.Client
	cli, err = util.DialSshRemote(&conf)
	if err != nil {
		return err
	} else {
		_ = cli.Close()
	}
	errChan := make(chan error, 1)
	readyChan := make(chan struct{}, 1)
	go func() {
		for ctx.Err() == nil {
			err := util.Main(ctx, remote, local, &conf, readyChan)
			if err != nil {
				if !errors.Is(err, context.Canceled) {
					log.Errorf("ssh forward failed err: %v", err)
				}
				select {
				case errChan <- err:
				default:
				}
			}
			time.Sleep(time.Second * 2)
		}
	}()
	log.Infof("wait jump to bastion host...")
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-readyChan:
		return nil
	case err = <-errChan:
		log.Errorf("ssh proxy err: %v", err)
		return err
	}
}

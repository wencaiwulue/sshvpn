package cmds

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"

	"github.com/containernetworking/cni/pkg/types"
	"github.com/spf13/cobra"
	"github.com/wencaiwulue/kubevpn/v2/pkg/driver"
	"github.com/wencaiwulue/kubevpn/v2/pkg/util"

	"github.com/wencaiwulue/tlstunnel/pkg/client"
	"github.com/wencaiwulue/tlstunnel/pkg/config"
	pkgutil "github.com/wencaiwulue/tlstunnel/pkg/util"
)

func CmdClient() *cobra.Command {
	var mode config.ProxyMode
	var stack config.StackType
	var extraCIDR []string
	var sshConf util.SshConfig

	cmd := &cobra.Command{
		Use:   "client",
		Short: "client to connect server",
		Long:  `client to connect server`,
		PreRun: func(cmd *cobra.Command, args []string) {
			if !pkgutil.IsAdmin() {
				pkgutil.RunWithElevated()
				os.Exit(0)
			}
			if util.IsWindows() {
				driver.InstallWireGuardTunDriver()
			}
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch mode {
			case config.ProxyModeFull, config.ProxyModeSplit:
			default:
				return fmt.Errorf("not support proxy mode: %s", mode)
			}
			var routes []types.Route
			for _, cidr := range extraCIDR {
				_, c, err := net.ParseCIDR(cidr)
				if err != nil {
					return fmt.Errorf("invalid cidr %s, err: %v", cidr, err)
				}
				routes = append(routes, types.Route{Dst: *c})
			}
			ctx, cancelFunc := context.WithCancel(cmd.Context())
			defer cancelFunc()
			go func() {
				signals := make(chan os.Signal)
				signal.Notify(signals, os.Kill, os.Interrupt)
				<-signals
				cancelFunc()
			}()
			defer driver.UninstallWireGuardTunDriver()
			return client.Connect(ctx, routes, sshConf, mode, stack)
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVar((*string)(&mode), "mode", string(config.ProxyModeFull), "Traffic mode, full: bypass all traffic, split: bypass traffic intelligent")
	cmd.Flags().IntVarP(&config.TCPPort, "tcp-port", "t", config.TCPPort, "The tcp port of remote linux server")
	cmd.Flags().IntVarP(&config.UDPPort, "udp-port", "u", config.UDPPort, "The udp port of remote linux server")
	cmd.Flags().StringVar((*string)(&stack), "stack", string(config.DualStack), string("Network stack. ["+config.SingleStackIPv4+"|"+config.SingleStackIPv6+"|"+config.DualStack+"]"))
	cmd.Flags().StringArrayVar(&extraCIDR, "extra-cidr", []string{}, "CIDR string, eg: --cidr 192.168.0.159/24 --cidr 192.168.1.160/32")
	addSshFlags(cmd, &sshConf)
	return cmd
}

func addSshFlags(cmd *cobra.Command, sshConf *util.SshConfig) {
	// for ssh jumper host
	cmd.Flags().StringVar(&sshConf.Addr, "ssh-addr", "", "Optional ssh jump server address to dial as <hostname>:<port>, eg: 127.0.0.1:22")
	cmd.Flags().StringVar(&sshConf.User, "ssh-username", "", "Optional username for ssh jump server")
	cmd.Flags().StringVar(&sshConf.Password, "ssh-password", "", "Optional password for ssh jump server")
	cmd.Flags().StringVar(&sshConf.Keyfile, "ssh-keyfile", "", "Optional file with private key for SSH authentication")
	cmd.Flags().StringVar(&sshConf.ConfigAlias, "ssh-alias", "", "Optional config alias with ~/.ssh/config for SSH authentication")
	cmd.Flags().StringVar(&sshConf.GSSAPIPassword, "gssapi-password", "", "GSSAPI password")
	cmd.Flags().StringVar(&sshConf.GSSAPIKeytabConf, "gssapi-keytab", "", "GSSAPI keytab file path")
	cmd.Flags().StringVar(&sshConf.GSSAPICacheFile, "gssapi-cache", "", "GSSAPI cache file path, use command `kinit -c /path/to/cache USERNAME@RELAM` to generate")
}

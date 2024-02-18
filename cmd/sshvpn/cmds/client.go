package cmds

import (
	"context"
	"os"
	"os/signal"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/wencaiwulue/kubevpn/v2/pkg/driver"
	"github.com/wencaiwulue/kubevpn/v2/pkg/util"

	"github.com/wencaiwulue/tlstunnel/pkg/client"
	"github.com/wencaiwulue/tlstunnel/pkg/config"
	pkgutil "github.com/wencaiwulue/tlstunnel/pkg/util"
)

func CmdClient() *cobra.Command {
	var mode config.ProxyType
	var pacPath string
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
			case config.ProxyTypeGlobe:
				//if len(remote) == 0 {
				//	log.Fatal("Globe mode, remote ip should not be empty")
				//}
			case config.ProxyTypePAC:
				if len(pacPath) == 0 {
					log.Fatal("PAC mode, PAC path should not be empty")
				}
			default:
				log.Fatal("Not support proxy mode " + mode)
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
			return client.Connect(ctx, extraCIDR, sshConf)
		},
		SilenceUsage: true,
	}
	cmd.Flags().StringVar((*string)(&mode), "mode", string(config.ProxyTypeGlobe), "Only support mode globe or pac")
	_ = cmd.Flags().MarkHidden("mode")
	cmd.Flags().StringVar(&pacPath, "pac", "", "The path of PAC, can be a url or local path")
	_ = cmd.Flags().MarkHidden("pac")
	cmd.Flags().IntVarP(&config.TCPPort, "tcp-port", "t", config.TCPPort, "The tcp port of remote linux server")
	cmd.Flags().IntVarP(&config.UDPPort, "udp-port", "u", config.UDPPort, "The udp port of remote linux server")
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

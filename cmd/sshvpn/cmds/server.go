package cmds

import (
	"context"
	"os"
	"os/exec"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/wencaiwulue/tlstunnel/pkg/config"
	"github.com/wencaiwulue/tlstunnel/pkg/server"
)

func CmdServer() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "server",
		Long:  `server`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancelFunc := context.WithCancel(cmd.Context())
			defer cancelFunc()
			go func() {
				signals := make(chan os.Signal)
				signal.Notify(signals, os.Kill, os.Interrupt)
				<-signals
				cancelFunc()
			}()
			_ = exec.Command("ulimit", "-HSn", "102400").Run()
			return server.Serve(ctx, config.TCPPort, config.UDPPort)
		},
		SilenceUsage: true,
	}
	cmd.Flags().IntVarP(&config.TCPPort, "tcp-port", "t", config.TCPPort, "server listen tcp port")
	cmd.Flags().IntVarP(&config.UDPPort, "udp-port", "u", config.UDPPort, "server listen udp port")
	return cmd
}

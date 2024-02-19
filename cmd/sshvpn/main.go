package main

import (
	"context"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/wencaiwulue/tlstunnel/cmd/sshvpn/cmds"
)

func main() {
	cmd := NewTunnelCommand()
	logrus.SetLevel(logrus.DebugLevel)
	cmd.AddCommand(
		cmds.CmdServer(),
		cmds.CmdClient(),
		cmds.CmdVersion(),
	)
	_ = cmd.ExecuteContext(context.Background())
}

func NewTunnelCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "sshvpn",
		Short: "connect to remote network",
		Long: `
      connect to remote network.
      `,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}
}

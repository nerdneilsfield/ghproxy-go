package cmd

import (
	"github.com/nerdneilsfield/ghproxy-go/pkg/ghproxy"
	"github.com/spf13/cobra"
)

var (
	host          string
	port          int
	proxyJsDelivr bool
)

func newRunCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "run",
		Short:        "ghproxy-go run",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			ghproxy.Run(host, port, proxyJsDelivr)
		},
	}

	cmd.Flags().StringVarP(&host, "host", "H", "0.0.0.0", "Host to listen on")
	cmd.Flags().IntVarP(&port, "port", "P", 8080, "Port to listen on")
	cmd.Flags().BoolVarP(&proxyJsDelivr, "proxy-jsdelivr", "J", false, "Proxy jsdelivr")
	return cmd
}

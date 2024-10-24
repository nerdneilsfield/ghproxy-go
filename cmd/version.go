package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

func newVersionCmd(version string, buildTime string, gitCommit string) *cobra.Command {
	return &cobra.Command{
		Use:          "version",
		Short:        "ghproxy-go version",
		Args:         cobra.NoArgs,
		SilenceUsage: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("ghproxy-go")
			fmt.Println("A reverse proxy for github resources")
			fmt.Println("Author: dengqi935@gmail.com")
			fmt.Println("Github: https://github.com/nerdneilsfield/ghproxy-go")
			fmt.Println("Wiki: https://nerdneilsfield.github.io/ghproxy-go/")
			fmt.Fprintf(cmd.OutOrStdout(), "ghproxy-go: %s\n", version)
			fmt.Fprintf(cmd.OutOrStdout(), "buildTime: %s\n", buildTime)
			fmt.Fprintf(cmd.OutOrStdout(), "gitCommit: %s\n", gitCommit)
			fmt.Fprintf(cmd.OutOrStdout(), "goVersion: %s\n", runtime.Version())
		},
	}
}

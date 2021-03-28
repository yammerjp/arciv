package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

const versionStr = "0.0.1"

var (
	versionCmd = &cobra.Command{
		Use: "version",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Println(versionStr)
		},
		Short: "Print binary version",
		Long:  "Print binary version",
		Args:  cobra.NoArgs,
	}
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

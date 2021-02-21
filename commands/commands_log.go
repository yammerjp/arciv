package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

var (
	logCmd = &cobra.Command{
		Use: "log",
		Run: logCommand,
	}
)

func logCommand(cmd *cobra.Command, args []string) {
	if err := logAction(); err != nil {
		Exit(err, 1)
	}
}

func logAction() (err error) {

	commitList, err := loadCommitListSelf()
	if err != nil {
		return err
	}
	for _, cId := range commitList {
		fmt.Println(cId)
	}

	return nil
}

func init() {
	RootCmd.AddCommand(logCmd)
}

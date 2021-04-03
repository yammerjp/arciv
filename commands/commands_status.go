package commands

import (
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use:   "status",
		Run:   statusCommand,
		Short: "Print difference of files from the latest commit.",
		Long:  "Print difference of files from the latest commit.",
		Args:  cobra.NoArgs,
	}
)

var debugOption bool = false

func statusCommand(cmd *cobra.Command, args []string) {
	if err := statusAction(); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(statusCmd)
	statusCmd.Flags().BoolVarP(&simplyPrinting, "simple", "m", false, "Print simply")
	statusCmd.Flags().BoolVarP(&runFastlyOption, "fast", "s", false, "Check fastly with checking timestamp, without checking file hash")
	statusCmd.Flags().BoolVarP(&debugOption, "debug", "b", false, "Debug print")
}

func statusAction() (err error) {
	nowCommit, err := createCommitStructure()
	if err != nil {
		return err
	}

	latestCommit, err := SelfRepo().LoadLatestCommit()
	if err != nil {
		return err
	}

	deleted, added := diffTags(latestCommit.Tags, nowCommit.Tags)
	printDiffs(deleted, added)
	return nil
}

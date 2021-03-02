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

func statusCommand(cmd *cobra.Command, args []string) {
	if err := statusAction(); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(statusCmd)
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

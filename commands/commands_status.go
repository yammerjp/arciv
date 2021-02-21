package commands

import (
	"github.com/spf13/cobra"
)

var (
	statusCmd = &cobra.Command{
		Use: "status",
		Run: statusCommand,
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
	paths, err := findPaths()
	if err != nil {
		return err
	}
	nowPhotos, err := takePhotos(paths)
	if err != nil {
		return err
	}

	commitListSelf, err := loadCommitListSelf()
	if err != nil {
		return err
	}
	latestCommitId := commitListSelf[len(commitListSelf)-1]
	latestCommitPhotos, err := loadCommit(latestCommitId)
	if err != nil {
		return err
	}
	deleted, added := diffPhotos(latestCommitPhotos, nowPhotos)
	printDiffs(deleted, added)
	return nil
}

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
	paths, err := findPaths(rootDir(), []string{".arciv"})
	if err != nil {
		return err
	}
	nowPhotos, err := takePhotos(paths)
	if err != nil {
		return err
	}

	timelineSelf, err := selfRepo.loadTimeline()
	if err != nil {
		return err
	}
	latestCommitId := timelineSelf[len(timelineSelf)-1]
	latestCommitPhotos, err := loadCommit(latestCommitId)
	if err != nil {
		return err
	}
	deleted, added := diffPhotos(latestCommitPhotos, nowPhotos)
	printDiffs(deleted, added)
	return nil
}

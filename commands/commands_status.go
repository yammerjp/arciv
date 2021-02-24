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
	selfRepo := SelfRepo()
	nowPhotos, err := takePhotosSelfRepo()
	if err != nil {
		return err
	}

	latestCommitId, err := selfRepo.LoadLatestCommitId()
	if err != nil {
		return err
	}

	latestCommitPhotos, err := selfRepo.LoadPhotos(latestCommitId)
	if err != nil {
		return err
	}

	deleted, added := diffPhotos(latestCommitPhotos, nowPhotos)
	printDiffs(deleted, added)
	return nil
}

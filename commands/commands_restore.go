package commands

import (
	"errors"
	"github.com/spf13/cobra"
)

var (
	restoreCmd = &cobra.Command{
		Use: "restore",
		Run: restoreCommand,
	}
)

var dryRun bool

func restoreCommand(cmd *cobra.Command, args []string) {
	if err := restoreAction(args); err != nil {
		Exit(err, 1)
	}
}

func restoreAction(args []string) (err error) {
	selfRepo := SelfRepo()
	// dryRun := false
  if len(args) != 2 {
		return errors.New("Usage: arciv restore [repository-name] [alias]")
  }
  repoName := args[0]
  commitAlias := args[1]

	// fetch remoteCommit
	remoteRepo, err := findRepo(repoName)
	if err != nil {
		return err
	}
	remoteCommit, err := remoteRepo.LoadCommitFromAlias(commitAlias)
	if err != nil {
		return err
	}

	// check no changes
	localCommit, err := createCommitStructure()
	if err != nil {
		return err
	}
	localLatestCommitId, err := selfRepo.LoadLatestCommitId()
	if err != nil {
		return err
	}
	if localCommit.Id[9:] != localLatestCommitId[9:] {
		return errors.New("Directory structure is not saved with latest commit")
	}

	// filter blob hashes to recieve
	localHashStrings, err := selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	for _, lPhoto := range localCommit.Photos {
		localHashStrings = append(localHashStrings, lPhoto.Hash.String())
	}
	var blobsToRecieve []Photo
	for _, rPhoto := range remoteCommit.Photos {
		if !isIncluded(localHashStrings, rPhoto.Hash.String()) {
			blobsToRecieve = append(blobsToRecieve, rPhoto)
		}
	}

	// download
	if dryRun {
		message("Dry run")
		for _, b := range blobsToRecieve {
			message("Download: " + b.Hash.String())
		}
		return nil
	}
	err = remoteRepo.ReceiveRemoteBlobs(blobsToRecieve)
	if err != nil {
		return err
	}

	// mv all local files to .arciv/blob
	err = stashPhotos(localCommit.Photos)
	if err != nil {
		return err
	}

	// rename and copy
	err = unstashPhotos(remoteCommit.Photos)

	// FIXME: Add the restored commit to local timeline
	return nil
}

func init() {
	RootCmd.AddCommand(restoreCmd)
  restoreCmd.PersistentFlags().BoolVarP(&dryRun, "dry-run", "d", false, "Show downloading files if you excute the subcommand 'restore'")
}

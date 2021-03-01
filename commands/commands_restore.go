package commands

import (
	"errors"
	"github.com/spf13/cobra"
)

var (
	restoreCmd = &cobra.Command{
		Use:   "restore <repository> <commit>",
		Run:   restoreCommand,
		Short: "Restore filles from the specified repository's commit.",
		Long:  "Restore filles from the specified repository's commit with downloading files that don't exist on local.",
		Args:  cobra.ExactArgs(2),
	}
)

var dryRunning bool
var forceExcution bool

func restoreCommand(cmd *cobra.Command, args []string) {
	if err := restoreAction(args[0], args[1]); err != nil {
		Exit(err, 1)
	}
}

func restoreAction(repoName, commitAlias string) (err error) {
	selfRepo := SelfRepo()

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

	if !forceExcution {
		localLatestCommitId, err := selfRepo.LoadLatestCommitId()
		if err != nil {
			return err
		}
		if localCommit.Id[9:] != localLatestCommitId[9:] {
			return errors.New("Directory structure is not saved with latest commit")
		}
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
	// FIXME: Check remote blobs exists?....

	// download
	if dryRunning {
		message("Show downloading files if you excute 'restore'.")
		for _, b := range blobsToRecieve {
			message("Download: " + b.Hash.String() + ", Will locate to: " + b.Path)
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
	restoreCmd.Flags().BoolVarP(&dryRunning, "dry-run", "d", false, "Show downloading files if you excute the subcommand 'restore'")
	restoreCmd.Flags().BoolVarP(&forceExcution, "force", "f", false, "Restore forcely even if files of the self repository is not commited")
}

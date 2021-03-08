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

	// filter blob hashes to receive
	localHashStrings, err := selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	for _, lTag := range localCommit.Tags {
		localHashStrings = append(localHashStrings, lTag.Hash.String())
	}
	var blobsToReceive []Tag
	for _, rTag := range remoteCommit.Tags {
		if !isIncluded(localHashStrings, rTag.Hash.String()) {
			blobsToReceive = append(blobsToReceive, rTag)
		}
	}
	// FIXME: Check remote blobs exists?....

	// download
	if dryRunning {
		message("Show downloading files if you excute 'restore'.")
		for _, tag := range blobsToReceive {
			messageStdin("download: " + tag.Hash.String() + ", will locate to: " + tag.Path)
		}
		return nil
	}
	err = remoteRepo.ReceiveRemoteBlobs(blobsToReceive)
	if err != nil {
		return err
	}

	// mv all local files to .arciv/blob
	err = stashTags(localCommit.Tags)
	if err != nil {
		return err
	}

	// rename and copy
	err = unstashTags(remoteCommit.Tags)
	if err != nil {
		return err
	}

	return selfRepo.AddCommit(remoteCommit)
}

func init() {
	RootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVarP(&dryRunning, "dry-run", "d", false, "Show downloading files if you excute the subcommand 'restore'")
	restoreCmd.Flags().BoolVarP(&forceExcution, "force", "f", false, "Restore forcely even if files of the self repository is not commited")
}

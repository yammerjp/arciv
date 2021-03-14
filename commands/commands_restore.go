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

var dryRunningOption bool
var forceExcutionOption bool

func restoreCommand(cmd *cobra.Command, args []string) {
	if err := restoreAction(args[0], args[1]); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVarP(&dryRunningOption, "dry-run", "d", false, "Show downloading files if you excute the subcommand 'restore'")
	restoreCmd.Flags().BoolVarP(&forceExcutionOption, "force", "f", false, "Restore forcely even if files of the self repository is not commited")
	restoreCmd.Flags().BoolVarP(&runFastlyOption, "fast", "s", false, "Check fastly with checking timestamp, without checking file hash")
}

func restoreAction(repoName, commitAlias string) (err error) {
	selfRepo, remoteRepo, localCommit, remoteCommit, err := loadReposAndCommits(repoName, commitAlias, runFastlyOption)
	if err != nil {
		return err
	}
	if !forceExcutionOption {
		localLatestCommitId, err := selfRepo.LoadLatestCommitId()
		if err != nil {
			return err
		}
		if localCommit.Id[9:] != localLatestCommitId[9:] {
			return errors.New("Directory structure is not saved with latest commit")
		}
	}

	localBlobs, err := selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	blobsToReceive := blobsShouldReceive(localBlobs, localCommit.Tags, remoteCommit.Tags)

	err = downloadBlobs(remoteRepo, blobsToReceive, dryRunningOption)
	if err != nil {
		return err
	}

	err = replaceAllTags(localCommit.Tags, remoteCommit.Tags)
	if err != nil {
		return err
	}

	return selfRepo.AddCommit(remoteCommit)
}

func loadReposAndCommits(repoName string, remoteCommitAlias string, localRunFastly bool) (localRepo Repository, remoteRepo Repository, localCommit Commit, remoteCommit Commit, err error) {
	remoteRepo, err = findRepo(repoName)
	if err != nil {
		return Repository{}, Repository{}, Commit{}, Commit{}, err
	}
	remoteCommit, err = remoteRepo.LoadCommitFromAlias(remoteCommitAlias)
	if err != nil {
		return Repository{}, Repository{}, Commit{}, Commit{}, err
	}
	localCommit, err = createCommitStructure(localRunFastly)
	if err != nil {
		return Repository{}, Repository{}, Commit{}, Commit{}, err
	}
	return SelfRepo(), remoteRepo, localCommit, remoteCommit, nil
}

func blobsShouldReceive(localBlobs []string, localTags []Tag, remoteTags []Tag) (blobsToReceive []Tag) {
	for _, lTag := range localTags {
		localBlobs = append(localBlobs, lTag.Hash.String())
	}
	for _, rTag := range remoteTags {
		if !isIncluded(localBlobs, rTag.Hash.String()) {
			blobsToReceive = append(blobsToReceive, rTag)
		}
	}
	return blobsToReceive
}

func downloadBlobs(remoteRepo Repository, blobs []Tag, dryRunning bool) error {
	// download
	if dryRunning {
		message("Show downloading files if you excute 'restore'.")
		for _, tag := range blobs {
			messageStdin("download: " + tag.Hash.String() + ", will locate to: " + tag.Path)
		}
		return nil
	}
	return remoteRepo.ReceiveRemoteBlobs(blobs)
}

func replaceAllTags(from []Tag, to []Tag) error {
	// mv all local files to .arciv/blob
	err := stashTags(from)
	if err != nil {
		return err
	}
	// rename and copy
	return unstashTags(to)
}

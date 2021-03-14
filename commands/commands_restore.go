package commands

import (
	"errors"
	"github.com/spf13/cobra"
	"strconv"
)

var (
	restoreCmd = &cobra.Command{
		Use:   "restore",
		Run:   restoreCommand,
		Short: "Restore filles from the specified repository's commit.",
		Long: `Restore filles from the specified repository's commit with downloading files that don't exist on local.
Example:
        arciv restore --repository repo-remote --commit a84bfc
          ... restore files immefiately from the commit 'a84bfc' of the repository 'repo-remote'
`,
		Args: cobra.NoArgs,
	}
)

var dryRunningOption bool
var forceExcutionOption bool
var requestOption bool
var validDaysStrOption string

var RunningFromRequestOption string
var RunningFromLatestRequestOption bool

func restoreCommand(cmd *cobra.Command, args []string) {
	if err := restoreAction(); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().BoolVarP(&dryRunningOption, "dry-run", "d", false, "Show downloading files if you excute the subcommand 'restore'")
	restoreCmd.Flags().BoolVarP(&forceExcutionOption, "force", "f", false, "Restore forcely even if files of the self repository is not commited")
	restoreCmd.Flags().BoolVarP(&runFastlyOption, "fast", "s", false, "Check fastly with checking timestamp, without checking file hash")

	restoreCmd.Flags().StringVarP(&repositoryNameOption, "repository", "r", "", "repository name")
	restoreCmd.Flags().StringVarP(&commitAliasOption, "commit", "c", "", "commit id")
	restoreCmd.Flags().BoolVarP(&requestOption, "request", "q", false, "Send request to restore archived files in AWS S3 Glacier Deep Archive")
	restoreCmd.Flags().StringVarP(&validDaysStrOption, "valid-days", "v", "3", "valid days of restored archive files")
	//restoreCmd.Flags().BoolVarP(&RunningFromLatestRequestOption, "run-latest-requested", "l", false, "Download and place files that was requested latestly")
	restoreCmd.Flags().StringVarP(&RunningFromRequestOption, "run-requested", "e", "", "Download and place files from restore-request")
}

func restoreAction() (err error) {
	if RunningFromRequestOption != "" {
		return restoreActionFromRequested(RunningFromRequestOption)
	}

	if repositoryNameOption == "" {
		return errors.New("Need to specify repository name")
	}
	if commitAliasOption == "" {
		return errors.New("Need to specify commit alias")
	}

	if requestOption {
		return restoreActionRequest()
	}
	return restoreActionImmediately()
}

func restoreActionImmediately() (err error) {
	remoteRepo, err := findRepo(repositoryNameOption)
	if err != nil {
		return err
	}
	remoteCommit, err := remoteRepo.LoadCommitFromAlias(commitAliasOption)
	if err != nil {
		return err
	}
	localCommit, err := createCommitStructure()
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}
	return downloadAndReplace(remoteRepo, localCommit, remoteCommit)
}

func restoreActionRequest() (err error) {
	validDaysI, err := strconv.ParseInt(validDaysStrOption, 10, 32)
	if err != nil {
		return err
	}
	validDays := int32(validDaysI)

	selfRepo := SelfRepo()
	localCommit, err := createCommitStructure()
	if err != nil {
		return err
	}
	remoteRepo, err := findRepo(repositoryNameOption)
	if err != nil {
		return err
	}
	remoteCommit, err := remoteRepo.LoadCommitFromAlias(commitAliasOption)
	if err != nil {
		return err
	}

	localBlobs, err := selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	blobsToReceive := blobsShouldReceive(localBlobs, localCommit.Tags, remoteCommit.Tags)
	if len(blobsToReceive) == 0 {
		message("Restore request is unnecessary. You can excute restore immediately.")
		return nil
	}
	blobs, err := remoteRepo.ReceiveRemoteBlobsRequest(blobsToReceive, validDays)
	if err != nil {
		return err
	}
	restoreReqeustId := timestamp2string(timestampNow())
	err = selfRepo.WriteRestoreRequest(restoreReqeustId, RestoreRequest{
		Repository: remoteRepo,
		ValidDays:  validDays,
		Commit:     remoteCommit,
		Blobs:      blobs,
	})
	if err != nil {
		return err
	}
	message("Sending request is success!\n restore-request:" + restoreReqeustId)
	return nil
}

func restoreActionFromRequested(restoreRequestIdAlias string) error {
	selfRepo := SelfRepo()
	localCommit, err := createCommitStructure()
	if err != nil {
		return err
	}

	rId, req, err := selfRepo.LoadRestoreRequest(restoreRequestIdAlias)
	if err != nil {
		return err
	}
	message("restore-request:" + rId)
	return downloadAndReplace(req.Repository, localCommit, req.Commit)
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

func downloadAndReplace(remoteRepo Repository, localCommit, remoteCommit Commit) error {
	selfRepo := SelfRepo()
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

	// download
	if dryRunningOption {
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

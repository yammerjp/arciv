package commands

import (
	"github.com/spf13/cobra"
)

var (
	storeCmd = &cobra.Command{
		Use:   "store <repository>",
		Run:   storeCommand,
		Short: "Store files from the self repository to another repository.",
		Long:  "Create a commit and send new blobs and timeline to another repository.",
		Args:  cobra.ExactArgs(1),
	}
)

func storeCommand(cmd *cobra.Command, args []string) {
	if err := storeAction(args[0]); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(storeCmd)
}

func storeAction(repoName string) (err error) {
	remoteRepo, err := findRepo(repoName)
	if err != nil {
		return err
	}

	commit, err := createCommitStructure()
	if err != nil {
		return err
	}
	err = SelfRepo().AddCommit(commit)
	if err != nil {
		return err
	}
	message("created commit '" + commit.Id + "'")

	remoteHashStrings, err := remoteRepo.FetchBlobHashes()
	if err != nil {
		return err
	}

	// send blobs not stored on remote repository
	var tagsToSend []Tag
	for _, tag := range commit.Tags {
		if !isIncluded(remoteHashStrings, tag.Hash.String()) {
			tagsToSend = append(tagsToSend, tag)
		}
	}
	err = remoteRepo.SendLocalBlobs(tagsToSend)
	if err != nil {
		return err
	}

	return remoteRepo.AddCommit(commit)
}

func isIncluded(strs []string, s string) bool {
	for _, str := range strs {
		if str == s {
			return true
		}
	}
	return false
}

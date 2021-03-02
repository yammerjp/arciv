package commands

import (
	"bytes"
	"github.com/spf13/cobra"
	"path/filepath"
)

var (
	unstashCmd = &cobra.Command{
		Use:   "unstash",
		Run:   unstashCommand,
		Short: "Unstash latest commit's files from .arciv/blob directory",
		Long:  "Move files from .arciv/blob directory to the self repository based on the file list of the latest commit",
		Args:  cobra.NoArgs,
	}
)

func unstashCommand(cmd *cobra.Command, args []string) {
	if err := unstashAction(); err != nil {
		Exit(err, 1)
	}
}

func unstashAction() (err error) {
	latestCommit, err := SelfRepo().LoadLatestCommit()
	if err != nil {
		return err
	}
	err = unstashTags(latestCommit.Tags)
	if err != nil {
		return err
	}
	return nil
}

func unstashTags(tags []Tag) (err error) {
	selfRepo := SelfRepo()
	// mkdir
	dirSet := make(map[string]struct{})
	for _, tag := range tags {
		dirSet[filepath.Dir(tag.Path)] = struct{}{}
	}
	for dir, _ := range dirSet {
		err = mkdirAll(dir)
		if err != nil {
			return err
		}
	}

	// copy or move
	for i, tag := range tags {
		from := selfRepo.Path + "/.arciv/blob/" + tag.Hash.String()
		to := selfRepo.Path + "/" + tag.Path

		// If different files point to a same blob,
		//  the blob is copied on the first (, second, and ...) time,
		//  and moved on the last time
		keepInBlobDir := false
		for j := i + 1; j < len(tags); j++ {
			if bytes.Compare(tags[j].Hash, tag.Hash) == 0 {
				keepInBlobDir = true
			}
		}

		var msg string
		if keepInBlobDir {
			err = copyFile(from, to)
			msg = "copied "
		} else {
			err = moveFile(from, to)
			msg = "moved "
		}

		if err != nil {
			return err
		}
		message(msg + from + " -> " + to)
	}
	return nil
}

func init() {
	RootCmd.AddCommand(unstashCmd)
}

package commands

import (
	"bytes"
	"errors"
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
	// Guard to check to be able to excute unstash with .arcv/blob list and tags
	blobs, err := SelfRepo().FetchBlobHashes()
	if err != nil {
		return err
	}
	for _, tag := range tags {
		if !isIncluded(blobs, tag.Hash.String()) {
			return errors.New("local blob is missing from commit")
		}
	}

	root := fileOp.rootDir()
	// mkdir
	dirSet := make(map[string]struct{})
	for _, tag := range tags {
		dirSet[filepath.Dir(tag.Path)] = struct{}{}
	}
	for dir, _ := range dirSet {
		err = fileOp.mkdirAll(dir)
		if err != nil {
			return err
		}
	}

	// copy or move
	for i, tag := range tags {
		from := root + "/.arciv/blob/" + tag.Hash.String()
		to := root + "/" + tag.Path

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
			err = fileOp.copyFile(from, to)
			msg = "copied "
		} else {
			err = fileOp.moveFile(from, to)
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
	unstashCmd.Flags().BoolVarP(&debugOption, "debug", "b", false, "Debug print")
}

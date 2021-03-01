package commands

import (
	"bytes"
	"github.com/spf13/cobra"
	"os"
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
	err = unstashPhotos(latestCommit.Photos)
	if err != nil {
		return err
	}
	return nil
}

func unstashPhotos(photos []Photo) (err error) {
	selfRepo := SelfRepo()
	// mkdir
	dirSet := make(map[string]struct{})
	for _, photo := range photos {
		dirSet[filepath.Dir(photo.Path)] = struct{}{}
	}
	for dir, _ := range dirSet {
		err = os.MkdirAll(dir, 0777)
		if err != nil {
			return err
		}
	}

	// copy or move
	for i, photo := range photos {
		from := selfRepo.Path + "/.arciv/blob/" + photo.Hash.String()
		to := selfRepo.Path + "/" + photo.Path

		// If different files point to a same blob,
		//  the blob copy on the first (, second, and ...) time,
		//  and the blob rename on the last time
		keepInBlobDir := false
		for j := i + 1; j < len(photos); j++ {
			if bytes.Compare(photos[j].Hash, photo.Hash) == 0 {
				keepInBlobDir = true
			}
		}

		var msg string
		if keepInBlobDir {
			err = copyFile(from, to)
			msg = "copied "
		} else {
			err = os.Rename(from, to)
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

package commands

import (
	"bytes"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

var (
	unstashCmd = &cobra.Command{
		Use: "unstash",
		Run: unstashCommand,
	}
)

func unstashCommand(cmd *cobra.Command, args []string) {
	if err := unstashAction(); err != nil {
		Exit(err, 1)
	}
}

func unstashAction() (err error) {
	photos, err := takePhotosSelfRepo()
	if err != nil {
		return err
	}
	err = unstashPhotos(photos)
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

	// remove local .arciv/blob/*
	localHashStrings, err := selfRepo.FetchBlobHashes()
	if err != nil {
		return err
	}
	for _, blob := range localHashStrings {
		err = os.Remove(selfRepo.Path + "/.arciv/blob/" + blob)
		if err != nil {
			return err
		}
		message("removed " + selfRepo.Path + "/.arciv/blob/" + blob)
	}
	return nil
}

func init() {
	RootCmd.AddCommand(unstashCmd)
}

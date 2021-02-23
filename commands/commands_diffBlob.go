package commands

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	diffBlobCmd = &cobra.Command{
		Use: "diff-blob",
		Run: diffBlobCommand,
	}
)

func diffBlobCommand(cmd *cobra.Command, args []string) {
	if err := diffBlobAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(diffBlobCmd)
}

func diffBlobAction(args []string) (err error) {
	if len(args) != 2 {
		return errors.New("Usage: arciv diff-blob [commit-id] [commit-id]")
	}
	timelineSelf, err := selfRepo.loadTimeline()
	if err != nil {
		return err
	}
	commitId0, err := findCommitId(args[0], timelineSelf)
	if err != nil {
		return err
	}
	commitId1, err := findCommitId(args[1], timelineSelf)
	if err != nil {
		return err
	}
	if commitId0 == commitId1 {
		return errors.New("Same commit")
	}
	photos0, err := selfRepo.loadPhotos(commitId0)
	if err != nil {
		return err
	}
	photos1, err := selfRepo.loadPhotos(commitId1)
	if err != nil {
		return err
	}

	deleted, added := diffHashes(photos0, photos1)
	printDiffHashes(deleted, added)

	return nil
}

func printDiffHashes(deleted, added []Photo) {
	for _, photo := range deleted {
		fmt.Println("\x1b[31m" + "- " + photo.Hash.String() + "\x1b[0m")
	}
	for _, photo := range added {
		fmt.Println("\x1b[32m" + "+ " + photo.Hash.String() + "\x1b[0m")
	}
}

func diffHashes(photosBefore, photosAfter []Photo) (deleted, added []Photo) {
	ib, ia := 0, 0
	for ib < len(photosBefore) && ia < len(photosAfter) {
		compared := bytes.Compare(photosBefore[ib].Hash, photosAfter[ia].Hash)
		if compared == 0 {
			ib++
			ia++
		} else if compared < 0 {
			deleted = append(deleted, photosBefore[ib])
			ib++
		} else {
			added = append(added, photosAfter[ia])
			ia++
		}
	}
	for _, photo := range photosBefore[ib:] {
		deleted = append(deleted, photo)
	}
	for _, photo := range photosAfter[ia:] {
		added = append(added, photo)
	}
	return deleted, added
}

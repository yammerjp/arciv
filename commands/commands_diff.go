package commands

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var (
	diffCmd = &cobra.Command{
		Use: "diff",
		Run: diffCommand,
	}
)

func diffCommand(cmd *cobra.Command, args []string) {
	if err := diffAction(args); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(diffCmd)
}

func diffAction(args []string) (err error) {
	selfRepo := SelfRepo()
	if len(args) != 2 {
		return errors.New("Usage: arciv diff [commit-id] [commit-id]")
	}
	commit0, err := selfRepo.LoadCommitFromAlias(args[0])
	if err != nil {
		return err
	}
	commit1, err := selfRepo.LoadCommitFromAlias(args[1])
	if err != nil {
		return err
	}
	if commit0.Id == commit1.Id {
		return errors.New("Same commit")
	}
	printDiffs(diffPhotos(commit0.Photos, commit1.Photos))
	return nil
}

func diffPhotos(photosBefore, photosAfter []Photo) (deleted []Photo, added []Photo) {
	ib, ia := 0, 0
	for ib < len(photosBefore) && ia < len(photosAfter) {
		compared := comparePhoto(photosBefore[ib], photosAfter[ia])
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
	for _, c := range photosBefore[ib:] {
		deleted = append(deleted, c)
	}
	for _, c := range photosAfter[ia:] {
		added = append(added, c)
	}
	return deleted, added
}

func printDiffs(deleted, added []Photo) {
	for _, dc := range deleted {
		// same hash
		idx := findPhotoIndex(added, dc, FIND_HASH|FIND_PATH)
		if idx != -1 {
			fmt.Printf("update: %s, hash: %s, timestamp: \x1b[31m%.8x\x1b[0m -> \x1b[32m%.8x\x1b[0m\n", dc.Path, dc.Hash.String(), dc.Timestamp, added[idx].Timestamp)
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		idx = findPhotoIndex(added, dc, FIND_HASH)
		if idx != -1 {
			fmt.Printf("rename: \x1b[31m%s\x1b[0m -> \x1b[32m%s\x1b[0m, hash: %s\n", dc.Path, added[idx].Path, dc.Hash.String())
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// same path, but not same hash
		idx = findPhotoIndex(added, dc, FIND_PATH)
		if idx != -1 {
			fmt.Printf("rewrite: %s, hash: \x1b[31m%s\x1b[0m -> \x1b[32m%s\x1b[0m\n", dc.Path, dc.Hash.String(), added[idx].Hash.String())
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// similar photo is not found
		fmt.Printf("\x1b[31mdeleted: %s, hash: %s\x1b[0m\n", dc.Path, dc.Hash.String())
	}
	// similar photo is not found
	for _, ac := range added {
		fmt.Printf("\x1b[32madded: %s, hash: %s\x1b[0m\n", ac.Path, ac.Hash.String())
	}
}

func printDiffsSimple(deleted, added []Photo) {
	for _, c := range deleted {
		fmt.Println("\x1b[31m" + "- " + c.String() + "\x1b[0m")
	}
	for _, c := range added {
		fmt.Println("\x1b[32m" + "+ " + c.String() + "\x1b[0m")
	}
}

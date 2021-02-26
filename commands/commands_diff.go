package commands

import (
	"errors"
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
			message("update: " + dc.Path + ", hash: " + dc.Hash.String() + ", timestamp: \x1b[31m" + timestamp2string(dc.Timestamp) + "\x1b[0m -> \x1b[32m" + timestamp2string(added[idx].Timestamp) + "\x1b[0m")
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		idx = findPhotoIndex(added, dc, FIND_HASH)
		if idx != -1 {
			message("rename: \x1b[31m" + dc.Path + "\x1b[0m -> \x1b[32m" + added[idx].Path + "\x1b[0m, hash: " + dc.Hash.String())
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// same path, but not same hash
		idx = findPhotoIndex(added, dc, FIND_PATH)
		if idx != -1 {
			message("rewrite: " + dc.Path + ", hash: \x1b[31m" + dc.Hash.String() + "\x1b[0m -> \x1b[32m" + added[idx].Hash.String() + "\x1b[0m")
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// similar photo is not found
		message("\x1b[31mdeleted: " + dc.Path + ", hash: " + dc.Hash.String() + "\x1b[0m")
	}
	// similar photo is not found
	for _, ac := range added {
		message("\x1b[32madded: " + ac.Path + ", hash: " + ac.Hash.String() + "\x1b[0m")
	}
}

func printDiffsSimple(deleted, added []Photo) {
	for _, c := range deleted {
		message("\x1b[31m" + "- " + c.String() + "\x1b[0m")
	}
	for _, c := range added {
		message("\x1b[32m" + "+ " + c.String() + "\x1b[0m")
	}
}

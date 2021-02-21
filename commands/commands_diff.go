package commands

import (
	"bufio"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"
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
	if len(args) != 2 {
		return errors.New("Usage: arciv diff [commit-id] [commit-id]")
	}
	commitListSelf, err := loadCommitListSelf()
	if err != nil {
		return err
	}
	cId0, err := findCommitId(args[0], commitListSelf)
	if err != nil {
		return err
	}
	cId1, err := findCommitId(args[1], commitListSelf)
	if err != nil {
		return err
	}
	if cId0 == cId1 {
		return errors.New("Same commit")
	}
	c0, err := loadCommit(cId0)
	if err != nil {
		return err
	}
	c1, err := loadCommit(cId1)
	if err != nil {
		return err
	}

	deleted, added := diffPhotos(c0, c1)
	printDiffs(deleted, added)
	return nil
}

func findCommitId(alias string, commitIds []string) (foundCId string, err error) {
	foundCId = ""
	if alias == "" {
		return "", errors.New("Empty commit id is spacified")
	}

	for _, cId := range commitIds {
		fullhit, sha256hit := strings.HasPrefix(cId, alias), strings.HasPrefix(cId[9:], alias)
		if !fullhit && !sha256hit {
			continue
		}
		if foundCId != "" {
			return "", errors.New("The alias refer to more than 1 commit")
		}
		foundCId = cId
	}
	if foundCId == "" {
		return "", errors.New("Commit is not found")
	}
	return foundCId, nil

}

func loadCommitListSelf() ([]string, error) {
	rootDir, err := findRoot()
	if err != nil {
		return []string{}, err
	}
	return loadCommitList(rootDir + "/.arciv/commit/self")
}

func loadCommitList(filepath string) ([]string, error) {
	var commits []string
	f, err := os.Open(filepath)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		commits = append(commits, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}
	return commits, nil
}

func loadCommit(commitId string) (photos []Photo, err error) {
	rootDir, err := findRoot()
	if err != nil {
		return []Photo{}, err
	}
	f, err := os.Open(rootDir + "/.arciv/list/" + commitId)
	if err != nil {
		return []Photo{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		photo, err := genPhoto(scanner.Text())
		if err != nil {
			return []Photo{}, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
}

func diffPhotos(commitListBefore, commitListAfter []Photo) (deleted []Photo, added []Photo) {
	ib, ia := 0, 0
	for ib < len(commitListBefore) && ia < len(commitListAfter) {
		compared := comparePhoto(commitListBefore[ib], commitListAfter[ia])
		if compared == 0 {
			ib++
			ia++
		} else if compared < 0 {
			deleted = append(deleted, commitListBefore[ib])
			ib++
		} else {
			added = append(added, commitListAfter[ia])
			ia++
		}
	}
	for _, c := range commitListBefore[ib:] {
		deleted = append(deleted, c)
	}
	for _, c := range commitListAfter[ia:] {
		added = append(added, c)
	}
	return deleted, added
}

func printDiffs(deleted, added []Photo) {
	for _, dc := range deleted {
		// same sha256
		idx := findPhotoIndex(added, "", dc.Sha256, 0, FIND_SHA256)
		if idx != -1 {
			fmt.Printf("rename: \x1b[31m%s\x1b[0m -> \x1b[32m%s\x1b[0m, sha256: %s\n", dc.Path, added[idx].Path, hex.EncodeToString(dc.Sha256))
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// same path, but not same sha256
		idx = findPhotoIndex(added, dc.Path, []byte{}, 0, FIND_PATH)
		if idx != -1 {
			fmt.Printf("rewrite: %s, sha256: \x1b[31m%s\x1b[0m -> \x1b[32m%s\x1b[0m\n", dc.Path, hex.EncodeToString(dc.Sha256), hex.EncodeToString(added[idx].Sha256))
			added = append(added[:idx], added[idx+1:]...)
			continue
		}
		// similar photo is not found
		fmt.Printf("\x1b[31mdeleted: %s, sha256: %s\x1b[0m\n", dc.Path, hex.EncodeToString(dc.Sha256))
	}
	// similar photo is not found
	for _, ac := range added {
		fmt.Printf("\x1b[32madded: %s, sha256: %s\x1b[0m\n", ac.Path, hex.EncodeToString(ac.Sha256))
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

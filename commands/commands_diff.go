package commands

import (
	"bufio"
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
	timelineSelf, err := loadTimeline("self")
	if err != nil {
		return err
	}
	cId0, err := findCommitId(args[0], timelineSelf)
	if err != nil {
		return err
	}
	cId1, err := findCommitId(args[1], timelineSelf)
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
		fullhit, hashhit := strings.HasPrefix(cId, alias), strings.HasPrefix(cId[9:], alias)
		if !fullhit && !hashhit {
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

func loadTimeline(repoName string) ([]string, error) {
	repo, err := findRepo(repoName)
	if err != nil {
		return []string{}, err
	}
	repoPath, err := repo.LocalPath()
	if err != nil {
		return []string{}, err
	}
	return loadLines(repoPath + "/.arciv/timeline")
}

func Exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func loadLines(filepath string) ([]string, error) {
	if !Exists(filepath) {
		file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return []string{}, err
		}
		file.Close()
	}
	var lines []string
	f, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return []string{}, err
	}
	return lines, nil
}

func loadCommit(commitId string) (photos []Photo, err error) {
	f, err := os.Open(rootDir() + "/.arciv/list/" + commitId)
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

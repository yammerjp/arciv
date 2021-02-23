package commands

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
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
	// fetch commit list
	// load photos of latest commit
	// read blob list
	// diff blob list and commit
	if len(args) != 2 {
		return errors.New("Usage: arciv diff-blob [commit-id] [commit-id]")
	}
	timelineSelf, err := loadTimelineSelf()
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

	root := rootDir()
	hashes0, err := loadHashes(root + "/.arciv/list/" + cId0)
	if err != nil {
		return err
	}
	hashes1, err := loadHashes(root + "/.arciv/list/" + cId1)
	if err != nil {
		return err
	}
	deleted, added := diffHashes(hashes0, hashes1)
	printDiffHashes(deleted, added)

	return nil
}

func printDiffHashes(deleted, added []Hash) {
	for _, b := range deleted {
		fmt.Println("\x1b[31m" + "- " + hex.EncodeToString(b) + "\x1b[0m")
	}
	for _, b := range added {
		fmt.Println("\x1b[32m" + "+ " + hex.EncodeToString(b) + "\x1b[0m")
	}
}

func loadHashes(path string) (hashes []Hash, err error) {
	f, err := os.Open(path)
	if err != nil {
		return []Hash{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 64 {
			return []Hash{}, errors.New("Failed to load hash from a file")
		}
		hash, err := hex2hash(line[:64])
		if err != nil {
			return []Hash{}, err
		}
		hashes = append(hashes, hash)
	}
	return hashes, nil
}

func diffHashes(hashesBefore, hashesAfter []Hash) (deleted []Hash, added []Hash) {
	ib, ia := 0, 0
	for ib < len(hashesBefore) && ia < len(hashesAfter) {
		compared := bytes.Compare(hashesBefore[ib], hashesAfter[ia])
		if compared == 0 {
			ib++
			ia++
		} else if compared < 0 {
			deleted = append(deleted, hashesBefore[ib])
			ib++
		} else {
			added = append(added, hashesAfter[ia])
			ia++
		}
	}
	for _, hash := range hashesBefore[ib:] {
		deleted = append(deleted, hash)
	}
	for _, hash := range hashesAfter[ia:] {
		added = append(added, hash)
	}
	return deleted, added
}

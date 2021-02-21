package commands

import (
	"crypto/sha256"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"sort"
)

var (
	commitCmd = &cobra.Command{
		Use: "commit",
		Run: commitCommand,
	}
)

func commitCommand(cmd *cobra.Command, args []string) {
	if err := commitAction(); err != nil {
		Exit(err, 1)
	}
}

func init() {
	RootCmd.AddCommand(commitCmd)
}

func commitAction() (err error) {
	paths, err := findPaths()
	if err != nil {
		return err
	}
	photos, err := takePhotos(paths)
	if err != nil {
		return err
	}
	c, err := createCommit(photos)
	fmt.Fprintln(os.Stderr, "created commit '"+c.Id+"'")

	return nil
}

func takePhotos(paths []string) ([]Photo, error) {
	var photos []Photo
	for _, path := range paths {
		photo, err := takePhoto(path)
		if err != nil {
			return []Photo{}, err
		}
		photos = append(photos, photo)
	}
	sort.Slice(photos, func(i, j int) bool {
		return comparePhoto(photos[i], photos[j]) < 0
	})
	return photos, nil
}

func takePhoto(path string) (Photo, error) {
	hash, err := calcFileHash(path)
	if err != nil {
		return Photo{}, err
	}
	timestamp, err := readTimestamp(path)
	if err != nil {
		return Photo{}, err
	}

	return Photo{
		Path:      path,
		Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

func calcFileHash(path string) (Hash, error) {
	hasher := sha256.New()
	f, err := os.Open(path)
	if err != nil {
		return []byte{}, err
	}
	defer f.Close()
	if _, err := io.Copy(hasher, f); err != nil {
		return []byte{}, err
	}
	return hasher.Sum(nil), nil
}

func readTimestamp(path string) (int64, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return 0, err
	}
	return fileInfo.ModTime().Unix(), nil
}

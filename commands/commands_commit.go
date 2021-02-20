package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
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

type Photo struct {
	Path      string
	Sha256    []byte
	Timestamp int64
}

func (photo Photo) toString() string {
	return hex.EncodeToString(photo.Sha256) + " " + fmt.Sprintf("%.8x", photo.Timestamp) + " " + photo.Path
}

func commitAction() (err error) {
	rootDir, err := findRoot()
	if err != nil {
		return err
	}
	paths, err := findPaths(rootDir)
	if err != nil {
		return err
	}
	photos, err := takePhotos(paths)
	if err != nil {
		return err
	}
	for _, photo := range photos {
		fmt.Println(photo.toString())
	}
	return nil
}

func findRoot() (string, error) {
	// find root directory (exist .arciv)
	// ex . current dir is /hoge/fuga/wara
	// search /hoge/fuga/wara/.arciv , and next /hoge/fuga/.arciv , and next /hoge/.arciv , and next /.arciv
	currentDir, _ := os.Getwd()
	for dir := currentDir; strings.LastIndex(dir, "/") != -1; dir = dir[:strings.LastIndex(dir, "/")] {
		if f, err := os.Stat(dir + "/.arciv"); !os.IsNotExist(err) && f.IsDir() {
			return dir, nil
		}
	}
	return "", errors.New(".arciv is not found")
}

func findPaths(rootDir string) ([]string, error) {
	var paths []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			paths = append(paths, path[len(rootDir)+1:])
			return nil
		}
		if info.Name() == ".arciv" {
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil {
		return []string{}, err
	}
	return paths, nil
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
		for k, _ := range photos[i].Sha256 {
			if photos[i].Sha256[k] == photos[j].Sha256[k] {
				continue
			}
			return photos[i].Sha256[k] < photos[j].Sha256[k]
		}
		return true
	})
	return photos, nil
}

func takePhoto(path string) (Photo, error) {
	sha256, err := sha256sum(path)
	if err != nil {
		return Photo{}, err
	}
	timestamp, err := readTimestamp(path)
	if err != nil {
		return Photo{}, err
	}

	return Photo{
		Path:      path,
		Sha256:    sha256,
		Timestamp: timestamp,
	}, nil
}

func sha256sum(path string) ([]byte, error) {
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

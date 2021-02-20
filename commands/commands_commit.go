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
  "time"
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

func (photo Photo) String() string {
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
  filepath := rootDir + "/.arciv/list/latest"
  err = writePhotos(photos, filepath)
  if err != nil {
    return err
  }
  commitId, err := createCommitId(filepath)
  if err != nil {
    return err
  }
  err = os.Rename(filepath, rootDir + "/.arciv/list/" + commitId)
  if err != nil {
    return err
  }
  err = addCommitList(commitId, rootDir)
  if err != nil {
    return err
  }
  fmt.Fprintln(os.Stderr, "created commit '" + commitId + "'")

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

func writePhotos(photos []Photo, filepath string) error {
  file, err := os.Create(filepath)
  if err != nil {
    return err
  }
  defer file.Close()
	for _, photo := range photos {
    file.WriteString(photo.String() + "\n")
	}
  return nil
}

func createCommitId(filepath string) (string, error) {
  commitSha256 ,err := sha256sum(filepath)
  if err != nil {
    return "", err
  }
  commitTime := time.Now().Unix()
  return fmt.Sprintf("%.8x", commitTime) + "-" + hex.EncodeToString(commitSha256), nil
}

func addCommitList(commitId string, rootDir string) error {
  filepath := rootDir + "/.arciv/commit/self"
  file, err := os.OpenFile(filepath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
  if err != nil {
    return err
  }
  defer file.Close()
  fmt.Fprintln(file, commitId)
  return nil
}

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

type File struct {
	Path      string
	Sha256    []byte
	Timestamp int64
}

func (file File) toString() string {
	return hex.EncodeToString(file.Sha256) + " " + fmt.Sprintf("%.8x", file.Timestamp) + " " + file.Path
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
	files, err := readFiles(paths)
	if err != nil {
		return err
	}
	for _, file := range files {
		fmt.Println(file.toString())
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

func readFiles(paths []string) ([]File, error) {
	var files []File
	for _, path := range paths {
		file, err := readFile(path)
		if err != nil {
			return []File{}, err
		}
		files = append(files, file)
	}
	sort.Slice(files, func(i, j int) bool {
		for k, _ := range files[i].Sha256 {
			if files[i].Sha256[k] == files[j].Sha256[k] {
				continue
			}
			return files[i].Sha256[k] < files[j].Sha256[k]
		}
		return true
	})
	return files, nil
}

func readFile(path string) (File, error) {
	sha256, err := sha256sum(path)
	if err != nil {
		return File{}, err
	}
	timestamp, err := readTimestamp(path)
	if err != nil {
		return File{}, err
	}

	return File{
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

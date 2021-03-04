package commands

import (
	"bufio"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func findPaths(root string, includeFile bool, includeDir bool) (relativePaths []string, err error) {
	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		isDir := info.IsDir()
		if isDir && !includeDir || !isDir && !includeFile {
			return nil
		}
		if len(root) >= len(path) {
			// exclude root directory
			return nil
		}
		relativePath := path[len(root)+1:]
		if isDir && relativePath == ".arciv" {
			return nil
		}
		if strings.HasPrefix(relativePath, ".arciv/") {
			return nil
		}
		// add relative path from root directory
		relativePaths = append(relativePaths, relativePath)
		return nil
	})
	if err != nil {
		return []string{}, err
	}
	return relativePaths, nil
}

func message(str string) {
	fmt.Fprintln(os.Stderr, str)
}

func messageStdin(str string) {
	fmt.Println(str)
}

type FileOp struct {
	copyFile      func(from, to string) error
	moveFile      func(from, to string) error
	removeFile    func(path string) error
	mkdirAll      func(path string) error
	hashFile      func(path string) (Hash, error)
	timestampFile func(path string) (int64, error)
	findFilePaths func(root string) ([]string, error)
	findDirPaths  func(root string) ([]string, error)
	writeLines    func(path string, lines []string) error
	loadLines     func(path string) ([]string, error)
	rootDir       func() string
}

var fileOp *FileOp

func init() {
	fileOp = &FileOp{
		copyFile: func(from, to string) error {
			w, err := os.Create(to)
			if err != nil {
				return err
			}
			defer w.Close()

			r, err := os.Open(from)
			if err != nil {
				return err
			}
			defer r.Close()

			_, err = io.Copy(w, r)
			return err
		},

		removeFile: func(path string) error {
			return os.Remove(path)
		},

		moveFile: func(from, to string) error {
			return os.Rename(from, to)
		},

		mkdirAll: func(path string) error {
			return os.MkdirAll(path, 0777)
		},

		hashFile: func(path string) (Hash, error) {
			hasher := sha256.New()
			f, err := os.Open(path)
			if err != nil {
				return Hash{}, err
			}
			defer f.Close()
			_, err = io.Copy(hasher, f)
			if err != nil {
				return Hash{}, err
			}
			return hasher.Sum(nil), nil
		},

		timestampFile: func(path string) (int64, error) {
			fileInfo, err := os.Stat(path)
			if err != nil {
				return 0, err
			}
			timestamp := fileInfo.ModTime().Unix()
			return timestamp, nil
		},

		findFilePaths: func(root string) ([]string, error) {
			return findPaths(root, true, false)
		},

		findDirPaths: func(root string) ([]string, error) {
			return findPaths(root, false, true)
		},

		rootDir: func() string {
			// find arciv's root directory (exist .arciv)
			// ex . current dir is /hoge/fuga/wara
			// search /hoge/fuga/wara/.arciv , and next /hoge/fuga/.arciv , and next /hoge/.arciv , and next /.arciv
			currentDir, _ := os.Getwd()
			for dir := currentDir; strings.LastIndex(dir, "/") != -1; dir = dir[:strings.LastIndex(dir, "/")] {
				if f, err := os.Stat(dir + "/.arciv"); !os.IsNotExist(err) && f.IsDir() {
					return dir
				}
			}
			Exit(errors.New(".arciv is not found"), 1)
			return ""
		},

		loadLines: func(path string) ([]string, error) {
			var lines []string
			f, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
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
		},

		writeLines: func(path string, lines []string) error {
			file, err := os.Create(path)
			if err != nil {
				return err
			}
			defer file.Close()

			fw := bufio.NewWriter(file)
			defer fw.Flush()
			for _, line := range lines {
				fmt.Fprintln(fw, line)
			}
			return nil
		},
	}
}

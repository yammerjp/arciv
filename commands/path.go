package commands

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var rootDir string

func init() {
	// find arciv's root directory (exist .arciv)
	// ex . current dir is /hoge/fuga/wara
	// search /hoge/fuga/wara/.arciv , and next /hoge/fuga/.arciv , and next /hoge/.arciv , and next /.arciv
	currentDir, _ := os.Getwd()
	for dir := currentDir; strings.LastIndex(dir, "/") != -1; dir = dir[:strings.LastIndex(dir, "/")] {
		if f, err := os.Stat(dir + "/.arciv"); !os.IsNotExist(err) && f.IsDir() {
			rootDir = dir
			return
		}
	}
	Exit(errors.New(".arciv is not found"), 1)
	return
}

func findPaths(root string, skipNames []string) ([]string, error) {
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		name := info.Name()
		for _, sname := range skipNames {
			if sname != name {
				continue
			}
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if !info.IsDir() {
			// add relative path from root directory
			paths = append(paths, path[len(root)+1:])
		}
		return nil
	})

	if err != nil {
		return []string{}, err
	}
	return paths, nil
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

func copyFile(from string, to string) error {
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
}

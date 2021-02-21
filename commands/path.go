package commands

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
)

func rootDir() string {
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
	return "" // don't call anytime
}

func findPaths() ([]string, error) {
	root := rootDir()
	var paths []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			// add relative path from arciv's root directory
			paths = append(paths, path[len(root)+1:])
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

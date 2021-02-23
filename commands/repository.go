package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Repository struct {
	Name string
	Path string
}

func (repository Repository) String() string {
	return repository.Name + " " + repository.Path
}

func (repository Repository) LocalPath() (string, error) {
	if !strings.HasPrefix(repository.Path, "file:///") {
		return "", errors.New("The repository does not have local path")
	}
	return repository.Path[len("file://"):], nil
}

func (repository Repository) AddTimeline(commit Commit) error {
	root, err := repository.LocalPath()
	if err != nil {
		return err
	}
	file, err := os.OpenFile(root+"/.arciv/timeline", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Fprintln(file, commit.Id)
	return nil
}

func (repository Repository) loadTimeline() ([]string, error) {
	repoPath, err := repository.LocalPath()
	if err != nil {
		return []string{}, err
	}
	return loadLines(repoPath + "/.arciv/timeline")
}

var selfRepo Repository

func init() {
	selfRepo = Repository{Name: "self", Path: "file://" + rootDir()}
}

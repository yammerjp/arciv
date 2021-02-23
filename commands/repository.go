package commands

import (
	"errors"
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

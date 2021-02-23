package commands

import (
	"bufio"
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

func (repository Repository) FilePath() (string, error) {
	if !strings.HasPrefix(repository.Path, "file:///") {
		return "", errors.New("The repository does not have local path")
	}
	return repository.Path[len("file://"):], nil
}

func (repository Repository) AddTimeline(commit Commit) error {
	root, err := repository.FilePath()
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
	repoPath, err := repository.FilePath()
	if err != nil {
		return []string{}, err
	}
	return loadLines(repoPath + "/.arciv/timeline")
}

func (repository Repository) WritePhotos(commit Commit) error {
	root, err := repository.FilePath()
	if err != nil {
		return err
	}
	os.MkdirAll(root+"/.arciv/list", 0777)
	file, err := os.Create(root + "/.arciv/list/" + commit.Id)
	if err != nil {
		return err
	}
	defer file.Close()

	fw := bufio.NewWriter(file)
	defer fw.Flush()
	for _, photo := range commit.Photos {
		fmt.Fprintln(fw, photo.String())
	}
	return nil
}

func (repository Repository) loadPhotos(commitId string) (photos []Photo, err error) {
	root, err := repository.FilePath()
	if err != nil {
		return []Photo{}, err
	}
	f, err := os.Open(root + "/.arciv/list/" + commitId)
	if err != nil {
		return []Photo{}, err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		photo, err := genPhoto(scanner.Text())
		if err != nil {
			return []Photo{}, err
		}
		photos = append(photos, photo)
	}
	return photos, nil
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

var selfRepo Repository

func init() {
	selfRepo = Repository{Name: "self", Path: "file://" + rootDir}
}

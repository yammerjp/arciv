package commands

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

type PathType int

const (
	PATH_FILE PathType = 0
)

type Repository struct {
	Name     string
	Path     string
	PathType PathType
}

func (repository Repository) String() string {
	if repository.PathType == PATH_FILE {
		return repository.Name + " file://" + repository.Path
	} else {
		Exit(errors.New("PathType Must Be PATH_FILE"), 1)
		return ""
	}
}

func (repository Repository) AddTimeline(commit Commit) error {
	if repository.PathType != PATH_FILE {
		return errors.New("Repository's PathType must be PATH_FILE")
	}

	// TODO: 既に存在していたら追記しない
	file, err := os.OpenFile(repository.Path+"/.arciv/timeline", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Fprintln(file, commit.Id)
	return nil
}

func (repository Repository) LoadTimeline() ([]string, error) {
	if repository.PathType != PATH_FILE {
		return []string{}, errors.New("Repository's PathType must be PATH_FILE")
	}

	return loadLines(repository.Path + "/.arciv/timeline")
}

func (repository Repository) WritePhotos(commit Commit) error {
	if repository.PathType != PATH_FILE {
		return errors.New("Repository's PathType must be PATH_FILE")
	}

	// TODO:既に同名のファイルが存在したら書き込む必要はない
	os.MkdirAll(repository.Path+"/.arciv/list", 0777)
	file, err := os.Create(repository.Path + "/.arciv/list/" + commit.Id)
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

func (repository Repository) LoadLatestCommitId() (string, error) {
	timeline, err := repository.LoadTimeline()
	if err != nil {
		return "", err
	}
	return timeline[len(timeline)-1], nil
}

func (repository Repository) LoadCommitFromAlias(alias string) (Commit, error) {
	timeline, err := repository.LoadTimeline()
	if err != nil {
		return Commit{}, err
	}
	commitId, err := findCommitId(alias, timeline)
	if err != nil {
		return Commit{}, err
	}
	return repository.LoadCommit(commitId)
}

func (repository Repository) LoadCommit(commitId string) (Commit, error) {
	photos, err := repository.LoadPhotos(commitId)
	if err != nil {
		return Commit{}, err
	}
	timestamp, err := genTimestamp(commitId[:8])
	if err != nil {
		return Commit{}, err
	}
	hash, err := hex2hash(commitId[9:])
	if err != nil {
		return Commit{}, err
	}
	return Commit{Id: commitId, Timestamp: timestamp, Hash: hash, Photos: photos}, nil
}

func (repository Repository) LoadPhotos(commitId string) (photos []Photo, err error) {
	if repository.PathType != PATH_FILE {
		return []Photo{}, errors.New("Repository's PathType must be PATH_FILE")
	}

	f, err := os.Open(repository.Path + "/.arciv/list/" + commitId)
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

func (repository Repository) FetchBlobHashes() ([]string, error) {
	if repository.PathType != PATH_FILE {
		return []string{}, errors.New("Repository's PathType must be PATH_FILE")
	}

	//   - .arciv/blob が無ければ掘る
	os.MkdirAll(repository.Path+"/.arciv/blob", 0777)
	//   - repository の .arciv/blob のファイル一覧を取得する
	return findPaths(repository.Path+"/.arciv/blob", []string{}, false)
}

func (repository Repository) sendLocalBlobs(photos []Photo) error {
	if repository.PathType != PATH_FILE {
		return errors.New("Repository's PathType must be PATH_FILE")
	}

	for _, photo := range photos {
		from := rootDir() + "/" + photo.Path
		to := repository.Path + "/.arciv/blob/" + photo.Hash.String()
		err := copyFile(from, to)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "copied %s -> %s\n", from, to)
	}
	return nil
}

func (repository Repository) ReceiveRemoteBlobs(photos []Photo) error {
	if repository.PathType != PATH_FILE {
		return errors.New("Repository's PathType must be PATH_FILE")
	}

	for _, photo := range photos {
		from := repository.Path + "/.arciv/blob/" + photo.Hash.String()
		to := rootDir() + "/.arciv/blob/" + photo.Hash.String()
		err := copyFile(from, to)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "copied %s -> %s\n", from, to)
	}
	return nil
}

func findCommitId(alias string, commitIds []string) (foundCId string, err error) {
	foundCId = ""
	if alias == "" {
		return "", errors.New("Empty commit id is spacified")
	}

	for _, cId := range commitIds {
		fullhit := strings.HasPrefix(cId, alias)
		hashhit := strings.HasPrefix(cId[9:], alias)
		if !fullhit && !hashhit {
			continue
		}
		if foundCId != "" {
			return "", errors.New("The alias refer to more than 1 commit")
		}
		foundCId = cId
	}
	if foundCId == "" {
		return "", errors.New("Commit is not found")
	}
	return foundCId, nil
}

func SelfRepo() Repository {
	return Repository{Name: "self", Path: rootDir(), PathType: PATH_FILE}
}

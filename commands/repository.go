package commands

import (
	"errors"
	"sort"
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
	}
	Exit(errors.New("PathType Must Be PATH_FILE"), 1)
	return ""
}

func (repository Repository) AddCommit(commit Commit) error {
	timeline, err := repository.LoadTimeline()
	if err != nil {
		return err
	}

	if isIncluded(timeline, commit.Id) {
		message("The commit " + commit.Id + " already exists in the timeline of the repository " + repository.Name)
		return nil
	}

	var baseCommit *Commit
	if len(timeline) > 0 {
		latestCommitId := timeline[len(timeline)-1]
		if latestCommitId[9:] == commit.Hash.String() {
			message("Committing is canceled. A commit that same directory structure already exists")
			return nil
		}
		c, err := repository.LoadCommit(latestCommitId)
		if err != nil {
			return err
		}
		baseCommit = &c
	}
	err = repository.WriteTags(commit, baseCommit)
	if err != nil {
		return err
	}
	return repository.WriteTimeline(append(timeline, commit.Id))
}

func (repository Repository) WriteTimeline(timeline []string) error {
	if repository.PathType == PATH_FILE {
		return fileOp.writeLines(repository.Path+"/.arciv/timeline", timeline)
	}
	return errors.New("Repository's PathType must be PATH_FILE")
}

func (repository Repository) LoadTimeline() ([]string, error) {
	if repository.PathType == PATH_FILE {
		return fileOp.loadLines(repository.Path + "/.arciv/timeline")
	}
	return []string{}, errors.New("Repository's PathType must be PATH_FILE")
}

func (repository Repository) WriteTags(commit Commit, base *Commit) error {
	var lines []string
	if base == nil {
		lines = []string{"#arciv-commit-atom"}
		for _, tag := range commit.Tags {
			lines = append(lines, tag.String())
		}
	} else {
		lines = []string{"#arciv-commit-extension from:" + base.Id}
		deleted, added := diffTags(base.Tags, commit.Tags)
		for _, c := range deleted {
			lines = append(lines, "- "+c.String())
		}
		for _, c := range added {
			lines = append(lines, "+ "+c.String())
		}
	}

	if repository.PathType == PATH_FILE {
		return fileOp.writeLines(repository.Path+"/.arciv/list/"+commit.Id, lines)
	}
	return errors.New("Repository's PathType must be PATH_FILE")
}

func (repository Repository) LoadTags(commitId string) (tags []Tag, err error) {
	var lines []string
	if repository.PathType == PATH_FILE {
		lines, err = fileOp.loadLines(repository.Path + "/.arciv/list/" + commitId)
		if err != nil {
			return []Tag{}, err
		}
	} else {
		return []Tag{}, errors.New("Repository's PathType must be PATH_FILE")
	}

	if strings.HasPrefix(lines[0], "#arciv-commit-extension from:") {
		if len(lines[0]) < 29+73 {
			return []Tag{}, errors.New("Length of the line '#arciv-commit-extension from:...' must be 102 or more")
		}
		return repository.LoadTagsFromExtension(lines[0][len("#arciv-commit-extension from:"):], lines[1:])
	}
	if !strings.HasPrefix(lines[0], "#arciv-commit-atom") && strings.HasPrefix(lines[0], "#") {
		return []Tag{}, errors.New("Unknow file type of a arciv tag list file")
	}

	for _, line := range lines[1:] {
		tag, err := str2Tag(line)
		if err != nil {
			return []Tag{}, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func (repository Repository) LoadTagsFromExtension(baseCommitId string, body []string) ([]Tag, error) {
	tags, err := repository.LoadTags(baseCommitId)
	if err != nil {
		return []Tag{}, err
	}
	for _, line := range body {
		if len(line) < 75 || string(line[1]) != " " {
			return []Tag{}, errors.New("Length of lines of a commit of extension tag list must be 75 or more")
		}

		tag, err := str2Tag(line[2:])
		if err != nil {
			return []Tag{}, err
		}

		switch string(line[0]) {
		case "-":
			idx := findTagIndex(tags, tag, FIND_HASH|FIND_PATH)
			if idx == -1 {
				return []Tag{}, errors.New("A tag specified by extension tag list is not found")
			}
			tags = append(tags[:idx], tags[idx+1:]...)
		case "+":
			tags = append(tags, tag)
		default:
			return []Tag{}, errors.New("Lines of a commit of extension tag list must be started with '+' or '-'")
		}
	}
	sort.Slice(tags, func(i, j int) bool {
		return compareTag(tags[i], tags[j]) < 0
	})

	return tags, nil
}

func (repository Repository) LoadLatestCommitId() (string, error) {
	timeline, err := repository.LoadTimeline()
	if err != nil {
		return "", err
	}
	if len(timeline) == 0 {
		return "", errors.New("Commit does not exists")
	}
	return timeline[len(timeline)-1], nil
}

func (repository Repository) LoadLatestCommit() (Commit, error) {
	id, err := repository.LoadLatestCommitId()
	if err != nil {
		return Commit{}, err
	}
	return repository.LoadCommit(id)
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
	if len(commitId) != 73 {
		return Commit{}, errors.New("Length of a commit id must be 73.")
	}
	tags, err := repository.LoadTags(commitId)
	if err != nil {
		return Commit{}, err
	}
	timestamp, err := str2timestamp(commitId[:8])
	if err != nil {
		return Commit{}, err
	}
	hash, err := hex2hash(commitId[9:])
	if err != nil {
		return Commit{}, err
	}
	return Commit{Id: commitId, Timestamp: timestamp, Hash: hash, Tags: tags}, nil
}

func (repository Repository) FetchBlobHashes() ([]string, error) {
	if repository.PathType == PATH_FILE {
		return fileOp.findFilePaths(repository.Path + "/.arciv/blob")
	}
	return []string{}, errors.New("Repository's PathType must be PATH_FILE")
}

// send from repository's root directory
func (repository Repository) SendLocalBlob(tag Tag) error {
	if repository.PathType == PATH_FILE {
		from := fileOp.rootDir() + "/" + tag.Path
		to := repository.Path + "/.arciv/blob/" + tag.Hash.String()
		return fileOp.copyFile(from, to)
	}
	return errors.New("Repository's PathType must be PATH_FILE")
}

// receive to .arciv/blob
func (repository Repository) ReceiveRemoteBlob(tag Tag) error {
	if repository.PathType == PATH_FILE {
		from := repository.Path + "/.arciv/blob/" + tag.Hash.String()
		to := fileOp.rootDir() + "/.arciv/blob/" + tag.Hash.String()
		return fileOp.copyFile(from, to)
	}
	return errors.New("Repository's PathType must be PATH_FILE")
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
	return Repository{Name: "self", Path: fileOp.rootDir(), PathType: PATH_FILE}
}

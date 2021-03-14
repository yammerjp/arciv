package commands

import (
	"errors"
	"sort"
	"strings"
)

const COMMIT_EXTENSION_DEPTH_MAX = 9

type Repository struct {
	Name     string
	Location RepositoryLocation
}

type RepositoryLocation interface {
	String() string
	writeLines(string, []string) error
	loadLines(string) ([]string, error)
	findFilePaths(string) ([]string, error)
	SendLocalBlobs([]Tag) error
	ReceiveRemoteBlobs([]Tag) error
	Init() error
}

func (repository Repository) String() string {
	return "name:" + repository.Name + " " + repository.Location.String()
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
		if c.Depth < COMMIT_EXTENSION_DEPTH_MAX {
			baseCommit = &c
		}
	}
	err = repository.WriteTags(commit, baseCommit)
	if err != nil {
		return err
	}
	return repository.WriteTimeline(append(timeline, commit.Id))
}

func (repository Repository) WriteTimeline(timeline []string) error {
	return repository.Location.writeLines(".arciv/timeline", timeline)
}

func (repository Repository) LoadTimeline() ([]string, error) {
	return repository.Location.loadLines(".arciv/timeline")
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
	err := repository.Location.writeLines(".arciv/list/"+commit.Id, lines)
	if err != nil {
		return err
	}
	lines = []string{"#arciv-timestamps of:" + commit.Id}
	for _, tag := range commit.Tags {
		if !tag.UsedTimestamp {
			return nil
		}
		lines = append(lines, tag.Hash.String()+" "+timestamp2string(tag.Timestamp))
	}
	return repository.Location.writeLines(".arciv/timestamps", lines)
}

func (repository Repository) LoadTimestamps(commitId string) ([]Tag, error) {
	lines, err := repository.Location.loadLines(".arciv/timestamps")
	if err != nil {
		return []Tag{}, err
	}
	if len(lines) == 0 {
		// cache is not exist
		return []Tag{}, nil
	}
	if !strings.HasPrefix(lines[0], "#arciv-timestamps of:") {
		return []Tag{}, errors.New("The first line of .arciv/timestamps must be started with '#arciv-timestamps of:")
	}
	if len(lines[0]) < len("#arciv-timestamps of:")+8+1+64 {
		return []Tag{}, errors.New("The first line of .arciv/timestamps is too short")
	}
	if lines[0][len("#arciv-timestamps of:"):len("#arciv-timestamps of:")+8+1+64] != commitId {
		// cache is old
		return []Tag{}, nil
	}
	var hashAndTimestamps []Tag
	for _, line := range lines[1:] {
		if len(line) < 64+1+8 {
			return []Tag{}, errors.New("a line in .arciv/timestamps is too short")
		}
		hash, err := hex2hash(line[:64])
		if err != nil {
			return []Tag{}, err
		}
		timestamp, err := str2timestamp(line[65 : 65+8])
		if err != nil {
			return []Tag{}, err
		}
		hashAndTimestamps = append(hashAndTimestamps, Tag{Hash: hash, Timestamp: timestamp, UsedTimestamp: true})
	}
	return hashAndTimestamps, nil
}

func (repository Repository) LoadTags(commitId string) (tags []Tag, depth int, err error) {
	hashAndTimestamps, err := repository.LoadTimestamps(commitId)
	if err != nil {
		return []Tag{}, 0, err
	}
	tags, depth, err = repository.loadTagsRecursive(commitId, 0)
	if err != nil {
		return []Tag{}, 0, err
	}
	for i, tag := range tags {
		timestampTagIndex := findTagIndex(hashAndTimestamps, Tag{Hash: tag.Hash}, FIND_HASH)
		if timestampTagIndex != -1 {
			tags[i].Timestamp = hashAndTimestamps[timestampTagIndex].Timestamp
			tags[i].UsedTimestamp = true
		}
	}
	return tags, depth, nil
}

func (repository Repository) loadTagsRecursive(commitId string, depth int) (tags []Tag, retDepth int, err error) {
	lines, err := repository.Location.loadLines(".arciv/list/" + commitId)
	if err != nil {
		return []Tag{}, 0, err
	}

	// #arciv-commit-atom
	if strings.HasPrefix(lines[0], "#arciv-commit-atom") {
		tags, err := loadTagsFromAtom(lines[1:])
		return tags, depth, err
	}
	// backward compatible
	if !strings.HasPrefix(lines[0], "#") {
		tags, err := loadTagsFromAtom(lines)
		return tags, depth, err
	}
	// #arciv-commit-extension
	if strings.HasPrefix(lines[0], "#arciv-commit-extension from:") {
		if len(lines[0]) < 29+73 {
			return []Tag{}, 0, errors.New("Length of the line '#arciv-commit-extension from:...' must be 102 or more")
		}
		commitIdFrom := lines[0][len("#arciv-commit-extension from:"):]
		tags, retDepth, err := repository.loadTagsRecursive(commitIdFrom, depth)
		if err != nil {
			return []Tag{}, 0, err
		}
		tags, err = loadTagsFromExtension(tags, lines[1:])
		return tags, retDepth + 1, err
	}
	return []Tag{}, 0, errors.New("Unknow file type of a arciv tag list file")
}

func loadTagsFromAtom(body []string) (tags []Tag, err error) {
	for _, line := range body {
		tag, err := str2Tag(line)
		if err != nil {
			return []Tag{}, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func loadTagsFromExtension(tags []Tag, body []string) ([]Tag, error) {
	for _, line := range body {
		if len(line) <= 2+64+1 || string(line[1]) != " " {
			return []Tag{}, errors.New("Length of lines of a commit of extension tag list must be 67 or more")
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
	tags, depth, err := repository.LoadTags(commitId)
	if err != nil {
		return Commit{}, err
	}
	sort.Slice(tags, func(i, j int) bool {
		return compareTag(tags[i], tags[j]) < 0
	})
	timestamp, err := str2timestamp(commitId[:8])
	if err != nil {
		return Commit{}, err
	}
	hash, err := hex2hash(commitId[9:])
	if err != nil {
		return Commit{}, err
	}
	return Commit{Id: commitId, Timestamp: timestamp, Hash: hash, Tags: tags, Depth: depth}, nil
}

func (repository Repository) FetchBlobHashes() (blobs []string, err error) {
	filenames, err := repository.Location.findFilePaths(".arciv/blob")
	if err != nil {
		return []string{}, err
	}
	for i, filename := range filenames {
		if len(filename) != 64 {
			filenames = append(filenames[:i], filenames[i+1:]...)
		}
	}
	return filenames, nil
}

// send from repository's root directory
func (repository Repository) SendLocalBlobs(tags []Tag) (err error) {
	return repository.Location.SendLocalBlobs(tags)
}

func (r Repository) ReceiveRemoteBlobsRequest(tags []Tag) (keysRequested []string, err error) {
	repositoryLocationS3, ok := r.Location.(RepositoryLocationS3)
	if !ok {
		return []string{}, errors.New("Repository.ReceiveRemoteBlobsRequest() is not succeeded with repository s3")
	}
	return repositoryLocationS3.ReceiveRemoteBlobsRequest(tags)
}

// receive to .arciv/blob
func (repository Repository) ReceiveRemoteBlobs(tags []Tag) (err error) {
	return repository.Location.ReceiveRemoteBlobs(tags)
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
	return Repository{Name: "self", Location: RepositoryLocationFile{Path: fileOp.rootDir()}}
}

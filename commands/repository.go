package commands

import (
	"errors"
	"sort"
	"strings"
)

type PathType int

const COMMIT_EXTENSION_DEPTH_MAX = 9

const (
	PATH_FILE PathType = 1
	PATH_S3   PathType = 2
)

type Repository struct {
	Name     string
	Path     string
	PathType PathType
}

func (repository Repository) String() string {

	if repository.PathType == PATH_FILE {
		return repository.Name + " file://" + repository.Path
	} else if repository.PathType == PATH_S3 {
		return repository.Name + " s3://" + repository.Path
	}
	Exit(errors.New("A repository with unknown PathType is not able to be stringified"), 1)
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
	if repository.PathType == PATH_FILE {
		return fileOp.writeLines(repository.Path+"/.arciv/timeline", timeline)
	}
	if repository.PathType == PATH_S3 {
		bucketName = repository.Path
		prepareS3BucketClient()
		return s3Op.writeLines(".arciv/timeline", timeline)
	}
	return errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
}

func (repository Repository) LoadTimeline() ([]string, error) {
	if repository.PathType == PATH_FILE {
		return fileOp.loadLines(repository.Path + "/.arciv/timeline")
	}
	if repository.PathType == PATH_S3 {
		bucketName = repository.Path
		prepareS3BucketClient()
		return s3Op.loadLines(".arciv/timeline")
	}
	return []string{}, errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
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
	if repository.PathType == PATH_S3 {
		bucketName = repository.Path
		prepareS3BucketClient()
		return s3Op.writeLines(".arciv/list/"+commit.Id, lines)
	}
	return errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
}

func (repository Repository) LoadTags(commitId string) (tags []Tag, depth int, err error) {
	return repository.loadTagsRecursive(commitId, 0)
}

func (repository Repository) loadTagsRecursive(commitId string, depth int) (tags []Tag, retDepth int, err error) {
	var lines []string
	if repository.PathType == PATH_FILE {
		lines, err = fileOp.loadLines(repository.Path + "/.arciv/list/" + commitId)
		if err != nil {
			return []Tag{}, 0, err
		}
	} else if repository.PathType == PATH_S3 {
		bucketName = repository.Path
		prepareS3BucketClient()
		lines, err = s3Op.loadLines(".arciv/list/" + commitId)
		if err != nil {
			return []Tag{}, 0, err
		}
	} else {
		return []Tag{}, 0, errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
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

func (repository Repository) FetchBlobHashes() ([]string, error) {
	if repository.PathType == PATH_FILE {
		return fileOp.findFilePaths(repository.Path + "/.arciv/blob")
	}
	if repository.PathType == PATH_S3 {
		bucketName = repository.Path
		prepareS3BucketClient()
		return s3Op.listBlobs()
	}
	return []string{}, errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
}

// send from repository's root directory
func (repository Repository) SendLocalBlobs(tags []Tag) (err error) {
	if repository.PathType == PATH_FILE {
		for _, tag := range tags {
			from := fileOp.rootDir() + "/" + tag.Path
			to := repository.Path + "/.arciv/blob/" + tag.Hash.String()
			err = fileOp.copyFile(from, to)
			if err != nil {
				return err
			}
			message("uploaded: " + tag.Hash.String() + ", " + tag.Path)
		}
		return nil
	}
	if repository.PathType == PATH_S3 {
		var fromPaths []string
		var blobNames []string
		for _, tag := range tags {
			fromPaths = append(fromPaths, fileOp.rootDir()+"/"+tag.Path)
			blobNames = append(blobNames, tag.Hash.String())
		}
		bucketName = repository.Path
		prepareS3BucketClient()
		return s3Op.sendBlobs(fromPaths, blobNames)
	}
	return errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
}

// receive to .arciv/blob
func (repository Repository) ReceiveRemoteBlobs(tags []Tag) (err error) {
	if repository.PathType == PATH_FILE {
		for _, tag := range tags {
			from := repository.Path + "/.arciv/blob/" + tag.Hash.String()
			to := fileOp.rootDir() + "/.arciv/blob/" + tag.Hash.String()
			err = fileOp.copyFile(from, to)
			if err != nil {
				return err
			}
			message("downloaded: " + tag.Hash.String() + ", will locate to: " + tag.Path)
		}
		return nil
	}
	if repository.PathType == PATH_S3 {
		// FIXME: receive restored files from deep archive
		return errors.New("Download blobs from AWS S3 is not implemented yet...\n Please download from Web console and place the files into .arciv/blob/")
		/*
			var toPaths []string
			var blobNames []string
			for _, tag := range tags {
				toPaths = append(toPaths, repository.Path+"/.arciv/blob/"+tag.Hash.String())
				blobNames = append(blobNames, tag.Hash.String())
			}
			bucketName = repository.Path
			prepareS3BucketClient()
			return s3Op.receiveBlobs(toPaths, blobNames)
		*/
	}
	return errors.New("Repository's PathType must be PATH_FILE or PATH_S3")
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

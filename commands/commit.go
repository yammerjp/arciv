package commands

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

type Commit struct {
	Id        string
	Timestamp int64
	Hash      Hash
	Tags      []Tag
}

func createCommit() (Commit, error) {
	selfRepo := SelfRepo()
	commit, err := createCommitStructure()
	if err != nil {
		return Commit{}, err
	}

	commitIds, err := selfRepo.LoadTimeline()
	if err != nil {
		return Commit{}, err
	}
	for _, cId := range commitIds {
		if cId[9:] == commit.Hash.String() {
			message("Committing is canceled. A commit that same directory structure already exists")
			return selfRepo.LoadCommit(cId)
		}
	}
	var baseCommit *Commit
	if len(commitIds) > 0 {
		c, err := selfRepo.LoadCommit(commitIds[len(commitIds)-1])
		if err != nil {
			return Commit{}, err
		}
		baseCommit = &c
	}

	err = selfRepo.WriteTags(commit, baseCommit)
	if err != nil {
		return Commit{}, err
	}
	err = selfRepo.AddTimeline(commit)
	if err != nil {
		return Commit{}, err
	}
	message("created commit '" + commit.Id + "'")
	return commit, nil
}

func createCommitStructure() (Commit, error) {
	// Tags
	tags, err := taggingsSelfRepo()
	if err != nil {
		return Commit{}, err
	}
	// Hash
	hasher := sha256.New()
	for _, tag := range tags {
		fmt.Fprintln(hasher, tag.String())
	}
	hash := Hash(hasher.Sum(nil))
	// Timestamp
	timestamp := time.Now().Unix()

	return Commit{
		Id:        timestamp2string(timestamp) + "-" + hash.String(),
		Timestamp: timestamp,
		Hash:      hash,
		Tags:      tags,
	}, nil
}

func taggingsSelfRepo() ([]Tag, error) {
	selfRepo := SelfRepo()
	paths, err := findPathsOfSelfRepo(true, false)
	if err != nil {
		return []Tag{}, err
	}

	var tags []Tag
	for _, path := range paths {
		tag, err := tagging(selfRepo.Path, path)
		if err != nil {
			return []Tag{}, err
		}
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool {
		return compareTag(tags[i], tags[j]) < 0
	})
	return tags, nil
}

func tagging(root, path string) (Tag, error) {
	// hash
	hasher := sha256.New()
	f, err := os.Open(root + "/" + path)
	if err != nil {
		return Tag{}, err
	}
	_, err = io.Copy(hasher, f)
	if err != nil {
		return Tag{}, err
	}
	hash := hasher.Sum(nil)
	f.Close()

	//timestamp
	fileInfo, err := os.Stat(root + "/" + path)
	if err != nil {
		return Tag{}, err
	}
	timestamp := fileInfo.ModTime().Unix()

	return Tag{
		Path:      path,
		Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

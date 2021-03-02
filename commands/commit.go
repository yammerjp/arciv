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

func createCommitStructure() (Commit, error) {
	// Tags
	selfRepo := SelfRepo()
	paths, err := findPathsOfSelfRepo(true, false)
	if err != nil {
		return Commit{}, err
	}
	var tags []Tag
	for _, path := range paths {
		tag, err := tagging(selfRepo.Path, path)
		if err != nil {
			return Commit{}, err
		}
		tags = append(tags, tag)
	}
	sort.Slice(tags, func(i, j int) bool {
		return compareTag(tags[i], tags[j]) < 0
	})

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

func tagging(root, relativePath string) (Tag, error) {
	// hash
	hasher := sha256.New()
	f, err := os.Open(root + "/" + relativePath)
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
	fileInfo, err := os.Stat(root + "/" + relativePath)
	if err != nil {
		return Tag{}, err
	}
	timestamp := fileInfo.ModTime().Unix()

	return Tag{
		Path:      relativePath,
		Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

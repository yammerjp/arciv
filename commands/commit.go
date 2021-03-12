package commands

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"time"
)

type Commit struct {
	Id        string
	Timestamp int64
	Hash      Hash
	Tags      []Tag
	Depth     int // memo chained commit depth. use in #arciv-commit-extension
}

var timestampNow func() int64

func init() {
	timestampNow = func() int64 {
		return time.Now().Unix()
	}
}

func createCommitStructure() (Commit, error) {
	// Tags
	root := fileOp.rootDir()
	paths, err := fileOp.findFilePaths(root)
	if err != nil {
		return Commit{}, err
	}
	var tags []Tag
	for _, path := range paths {
		tag, err := tagging(root, path)
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
	timestamp := timestampNow()

	return Commit{
		Id:        timestamp2string(timestamp) + "-" + hash.String(),
		Timestamp: timestamp,
		Hash:      hash,
		Tags:      tags,
		Depth:     0,
	}, nil
}

func tagging(root, relativePath string) (Tag, error) {
	path := root + "/" + relativePath
	// hash
	hash, err := fileOp.hashFile(path)
	if err != nil {
		return Tag{}, err
	}

	//timestamp
	timestamp, err := fileOp.timestampFile(path)
	if err != nil {
		return Tag{}, err
	}

	return Tag{
		Path:      relativePath,
		Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

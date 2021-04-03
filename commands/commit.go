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

var runFastlyOption bool
var timestampNow func() int64

func init() {
	timestampNow = func() int64 {
		return time.Now().Unix()
	}
}

func createCommitStructure() (c Commit, err error) {
	root := fileOp.rootDir()
	// Tags
	paths, err := fileOp.findFilePaths(root)
	if err != nil {
		return Commit{}, err
	}
	var tags []Tag
	for _, path := range paths {
		tag, err := tagging(root, path, !runFastlyOption)
		if err != nil {
			return Commit{}, err
		}
		tags = append(tags, tag)
	}
	if runFastlyOption {
		latestCommit, err := SelfRepo().LoadLatestCommit()
		if err != nil {
			return Commit{}, err
		}
		for i, tag := range tags {
			lci := findTagIndex(latestCommit.Tags, tag, FIND_PATH|FIND_TIMESTAMP)
			if lci != -1 && tag.UsedTimestamp {
				// path and timestamp is same as latest commit's one
				//   hash will be same as latest commit's one (fast mode)
				tags[i] = latestCommit.Tags[lci]
				continue
			}
			hash, err := fileOp.hashFile(root + "/" + tag.Path)
			if err != nil {
				return Commit{}, err
			}
			tags[i].Hash = hash
			tags[i].UsedHash = true
		}
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

func tagging(root, relativePath string, withHashing bool) (tag Tag, err error) {
	path := root + "/" + relativePath
	// hash

	var hash Hash
	if withHashing {
		hash, err = fileOp.hashFile(path)
		if err != nil {
			return Tag{}, err
		}
		if debugOption {
			message("(sha256) " + hash.String() + " " + path)
		}
	}

	//timestamp
	timestamp, err := fileOp.timestampFile(path)
	if err != nil {
		return Tag{}, err
	}

	return Tag{
		Path:          relativePath,
		Hash:          hash,
		Timestamp:     timestamp,
		UsedTimestamp: true,
		UsedHash:      withHashing,
	}, nil
}

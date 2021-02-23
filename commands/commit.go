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
	Photos    []Photo
}

func createCommit() (Commit, error) {
	// Photos
	photos, err := takePhotosSelfRepo()
	if err != nil {
		return Commit{}, err
	}
	// Hash
	hasher := sha256.New()
	for _, photo := range photos {
		fmt.Fprintln(hasher, photo.String())
	}
	hash := Hash(hasher.Sum(nil))
	// Timestamp
	timestamp := time.Now().Unix()

	commit := Commit{
		Id:        fmt.Sprintf("%.8x", timestamp) + "-" + hash.String(),
		Timestamp: timestamp,
		Hash:      hash,
		Photos:    photos,
	}

	commitIds, err := selfRepo.LoadTimeline()
	if err != nil {
		return Commit{}, err
	}
	for _, cId := range commitIds {
		if cId[9:] == commit.Hash.String() {
			fmt.Fprintln(os.Stderr, "Does not commit. A commit that same directory structure already exists")
			return selfRepo.LoadCommit(cId)
		}
	}

	err = selfRepo.WritePhotos(commit)
	if err != nil {
		return Commit{}, err
	}
	err = selfRepo.AddTimeline(commit)
	if err != nil {
		return Commit{}, err
	}
	fmt.Fprintln(os.Stderr, "created commit '"+commit.Id+"'")
	return commit, nil
}

func takePhotosSelfRepo() ([]Photo, error) {
	paths, err := findPaths(rootDir, []string{".arciv"})
	if err != nil {
		return []Photo{}, err
	}

	var photos []Photo
	for _, path := range paths {
		photo, err := takePhoto(rootDir, path)
		if err != nil {
			return []Photo{}, err
		}
		photos = append(photos, photo)
	}
	sort.Slice(photos, func(i, j int) bool {
		return comparePhoto(photos[i], photos[j]) < 0
	})
	return photos, nil
}

func takePhoto(root, path string) (Photo, error) {
	// hash
	hasher := sha256.New()
	f, err := os.Open(root + "/" + path)
	if err != nil {
		return Photo{}, err
	}
	_, err = io.Copy(hasher, f)
	if err != nil {
		return Photo{}, err
	}
	hash := hasher.Sum(nil)
	f.Close()

	//timestamp
	fileInfo, err := os.Stat(root + "/" + path)
	if err != nil {
		return Photo{}, err
	}
	timestamp := fileInfo.ModTime().Unix()

	return Photo{
		Path:      path,
		Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

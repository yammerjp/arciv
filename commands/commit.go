package commands

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"os"
	"time"
)

type Commit struct {
	Id        string
	Timestamp int64
	Hash      Hash
	Photos    []Photo
}

func createCommit() (Commit, error) {
	paths, err := findPaths(rootDir(), []string{".arciv"})
	if err != nil {
		return Commit{}, err
	}
	photos, err := takePhotos(paths)
	if err != nil {
		return Commit{}, err
	}
	commit, err := createCommitStruct(photos)
	fmt.Fprintln(os.Stderr, "created commit '"+commit.Id+"'")
	return commit, nil
}

func createCommitStruct(photos []Photo) (Commit, error) {
	hash := calcHash(photos)
	timestamp := time.Now().Unix()
	c := Commit{
		Id:        fmt.Sprintf("%.8x", timestamp) + "-" + hash.String(),
		Timestamp: timestamp,
		Hash:      hash,
		Photos:    photos,
	}
	err := c.WritePhotosSelf()
	if err != nil {
		return Commit{}, err
	}
	err = selfRepo.AddTimeline(c)
	if err != nil {
		return Commit{}, err
	}
	return c, nil
}

func calcHash(photos []Photo) Hash {
	hasher := sha256.New()
	for _, photo := range photos {
		fmt.Fprintln(hasher, photo.String())
	}
	return hasher.Sum(nil)
}

func (commit Commit) WritePhotosSelf() error {
	return commit.WritePhotos(rootDir())
}
func (commit Commit) WritePhotos(root string) error {
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

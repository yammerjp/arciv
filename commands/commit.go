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

func createCommit(photos []Photo) (Commit, error) {
	hash := calcHash(photos)
	timestamp := time.Now().Unix()
	c := Commit{
		Id:        fmt.Sprintf("%.8x", timestamp) + "-" + hash.String(),
		Timestamp: timestamp,
		Hash:      hash,
		Photos:    photos,
	}
	err := c.WritePhotos()
	if err != nil {
		return Commit{}, err
	}
	err = c.AddCommitListSelf()
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

func (commit Commit) AddCommitListSelf() error {
	file, err := os.OpenFile(rootDir()+"/.arciv/commit/self", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	defer file.Close()
	fmt.Fprintln(file, commit.Id)
	return nil
}

func (commit Commit) WritePhotos() error {
	file, err := os.Create(rootDir() + "/.arciv/list/" + commit.Id)
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

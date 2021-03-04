package commands

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Tag struct {
	Path      string
	Hash      Hash
	Timestamp int64
}

func (tag Tag) String() string {
	return tag.Hash.String() + " " + timestamp2string(tag.Timestamp) + " " + tag.Path
}

func str2Tag(line string) (Tag, error) {
	// 64...Hash, 1...space, 8...timestamp, 1...space
	if len(line) <= 64+1+8+1 {
		return Tag{}, errors.New("The length of Tag's line must be more than 74")
	}
	hash, err := hex2hash(line[:64])
	if err != nil {
		return Tag{}, err
	}
	timestamp, err := str2timestamp(line[65:73])
	if err != nil {
		return Tag{}, err
	}
	return Tag{
		Path:      line[74:],
		Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

func str2timestamp(str string) (int64, error) {
	return strconv.ParseInt(str, 16, 64)
}

func timestamp2string(t int64) string {
	return fmt.Sprintf("%.8x", t)
}

// priority  hash > timestamp > path
func compareTag(p0, p1 Tag) int {
	compared := bytes.Compare(p0.Hash, p1.Hash)
	if compared != 0 {
		return compared
	}
	diff := p0.Timestamp - p1.Timestamp
	if diff != 0 {
		if diff < 0 {
			return -1
		} else {
			return 1
		}
	}
	return strings.Compare(p0.Path, p1.Path)
}

type FindField int

const (
	FIND_PATH      FindField = 0x0001
	FIND_HASH      FindField = 0x0010
	FIND_TIMESTAMP FindField = 0x0100
)

func findTagIndex(tags []Tag, comparingTag Tag, flag FindField) int {
	for i, p := range tags {
		if (flag&FIND_HASH) != 0000 && bytes.Compare(p.Hash, comparingTag.Hash) != 0 {
			continue
		}
		if (flag&FIND_TIMESTAMP) != 0000 && p.Timestamp != comparingTag.Timestamp {
			continue
		}
		if (flag&FIND_PATH) != 0x0000 && p.Path != comparingTag.Path {
			continue
		}
		return i
	}
	return -1
}

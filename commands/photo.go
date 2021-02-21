package commands

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Hash []byte

func (hash Hash) String() string {
  return hex.EncodeToString(hash)
}

func hex2hash(hexStr string) (Hash, error) {
  return hex.DecodeString(hexStr)
}

type Photo struct {
	Path      string
  Hash      Hash
	Timestamp int64
}

func (photo Photo) String() string {
	return photo.Hash.String() + " " + fmt.Sprintf("%.8x", photo.Timestamp) + " " + photo.Path
}

func genPhoto(line string) (Photo, error) {
	// 64...Hash, 1...space, 8...timestamp, 1...space
	if len(line) <= 64+1+8+1 {
		return Photo{}, errors.New("The length of Photo's line must be more than 74")
	}
	hash, err := hex2hash(line[:64])
	if err != nil {
		return Photo{}, err
	}
	timestamp, err := strconv.ParseInt(line[65:73], 16, 64)
	if err != nil {
		return Photo{}, err
	}
	return Photo{
		Path:      line[74:],
    Hash:      hash,
		Timestamp: timestamp,
	}, nil
}

func comparePhoto(p0, p1 Photo) int {
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
	FIND_PATH      FindField = 0x0000
	FIND_HASH    FindField = 0x0010
	FIND_TIMESTAMP FindField = 0x0100
)

func findPhotoIndex(photos []Photo, path string, hash []byte, timestamp int64, flag FindField) int {
	for i, p := range photos {
		if (flag&FIND_HASH) == FIND_HASH && bytes.Compare(p.Hash, hash) != 0 {
			continue
		}
		if (flag&FIND_TIMESTAMP) == FIND_TIMESTAMP && p.Timestamp != timestamp {
			continue
		}
		if (flag&FIND_PATH) == FIND_PATH && p.Path == path {
			continue
		}
		return i
	}
	return -1
}

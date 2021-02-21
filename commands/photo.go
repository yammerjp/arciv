package commands

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type Photo struct {
	Path      string
	Sha256    []byte
	Timestamp int64
}

func (photo Photo) String() string {
	return hex.EncodeToString(photo.Sha256) + " " + fmt.Sprintf("%.8x", photo.Timestamp) + " " + photo.Path
}

func genPhoto(line string) (Photo, error) {
	// 64...sha256, 1...space, 8...timestamp, 1...space
	if len(line) <= 64+1+8+1 {
		return Photo{}, errors.New("The length of Photo's line must be more than 74")
	}
	sha256, err := hex.DecodeString(line[:64])
	if err != nil {
		return Photo{}, err
	}
	timestamp, err := strconv.ParseInt(line[65:73], 16, 64)
	if err != nil {
		return Photo{}, err
	}
	return Photo{
		Path:      line[74:],
		Sha256:    sha256,
		Timestamp: timestamp,
	}, nil
}

func comparePhoto(p0, p1 Photo) int {
	compared := bytes.Compare(p0.Sha256, p1.Sha256)
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
	FIND_SHA256    FindField = 0x0010
	FIND_TIMESTAMP FindField = 0x0100
)

func findPhotoIndex(photos []Photo, path string, sha256 []byte, timestamp int64, flag FindField) int {
	for i, p := range photos {
		if (flag&FIND_SHA256) == FIND_SHA256 && bytes.Compare(p.Sha256, sha256) != 0 {
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

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

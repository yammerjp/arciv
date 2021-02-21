package commands

import (
	"encoding/hex"
)

type Hash []byte

func (hash Hash) String() string {
	return hex.EncodeToString(hash)
}

func hex2hash(hexStr string) (Hash, error) {
	return hex.DecodeString(hexStr)
}

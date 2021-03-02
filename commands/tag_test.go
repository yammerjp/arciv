package commands

import (
	"bytes"
	"testing"
)

func TestTag(t *testing.T) {
	tagPath := ".git/hooks/applypatch-msg.sample"
	var tagHash Hash = []byte{0x02, 0x23, 0x49, 0x7a, 0x0b, 0x8b, 0x03, 0x3a, 0xa5, 0x8a, 0x3a, 0x52, 0x1b, 0x86, 0x29, 0x86, 0x93, 0x86, 0xcf, 0x7a, 0xb0, 0xe2, 0xf1, 0x01, 0x96, 0x3d, 0x32, 0x8a, 0xa6, 0x21, 0x93, 0xf7}
	var tagTimestamp int64 = 0x6030cc8d
	tag := Tag{
		Path:      tagPath,
		Hash:      tagHash,
		Timestamp: tagTimestamp,
	}
	tagString := "0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7 6030cc8d .git/hooks/applypatch-msg.sample"

	t.Run("Tag{}.String()", func(t *testing.T) {
		got := tag.String()
		want := tagString
		if got != want {
			t.Errorf("tag.String() = %s, want %s", got, want)
		}
	})

	t.Run("str2Tag()", func(t *testing.T) {
		got, err := str2Tag(tagString)
		if err != nil {
			t.Errorf("str2Tag(%s) return an error %s", tagString, err)
		}
		if got.Path != tagPath {
			t.Errorf("str2Tag(%s) = {Path: %s ...}, want %s", tagString, got.Path, tagPath)
		}
		if bytes.Compare(got.Hash, tagHash) != 0 {
			t.Errorf("str2Tag(%s) = {Hash: %s ...}, want %s", tagString, got.Hash.String(), tagHash.String())
		}
		if got.Timestamp != tagTimestamp {
			t.Errorf("str2Tag(%s) = {Timestamp: %.8x ...}, want %.8x", tagString, got.Timestamp, tagTimestamp)
		}
	})
}
package commands

import (
	"bytes"
	"testing"
)

func hashing(hexStr string) Hash {
	hash, err := hex2hash(hexStr)
	if err != nil {
		panic(err)
	}
	return hash
}

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

	// func (tag Tag) String() string
	t.Run("Tag{}.String()", func(t *testing.T) {
		got := tag.String()
		want := tagString
		if got != want {
			t.Errorf("tag.String() = %s, want %s", got, want)
		}
	})

	// func str2Tag(line string) (Tag, error)
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

	// func str2timestamp(str string) (int64, error)
	// func timestamp2string(t int64) string

	// func compareTag(p0, p1 Tag) int
	// compareTag(p0, p1) ... p0 - p1
	tag = Tag{Path: ".git/hooks/applypatch-msg.sample", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x6030cc8d}
	// L ... long, s ... short
	tagLpath := Tag{Path: ".hgit/hooks/applypatch-msg.sample", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x6030cc8d}
	tagSpath := Tag{Path: ".git/hooks", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x6030cc8d}

	tagLhash := Tag{Path: ".git/hooks/applypatch-msg.sample", Hash: hashing("ffffff7a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x6030cc8d}
	tagShash := Tag{Path: ".git/hooks/applypatch-msg.sample", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa6000000"), Timestamp: 0x6030cc8d}

	tagLtime := Tag{Path: ".git/hooks/applypatch-msg.sample", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0xffffff8d}
	tagStime := Tag{Path: ".git/hooks/applypatch-msg.sample", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x60000000}

	tagLhashStime := Tag{Path: ".git/hooks/applypatch-msg.sample", Hash: hashing("ffffff7a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x60000000}
	tagStimeLpath := Tag{Path: ".hgit/hooks/applypatch-msg.sample", Hash: hashing("0223497a0b8b033aa58a3a521b8629869386cf7ab0e2f101963d328aa62193f7"), Timestamp: 0x60000000}

	// comparing ... tag and others
	t.Run("compareTag(p0, p1 Tag) int", func(t *testing.T) {
		got := compareTag(tag, tag)
		if got != 0 {
			t.Errorf("compareTag(tag, tag) = %d, want 0", got)
		}
		got = compareTag(tag, tagLpath)
		if got != -1 {
			t.Errorf("compareTag(tag, tagLpath) = %d, want -1", got)
		}
		got = compareTag(tag, tagSpath)
		if got != 1 {
			t.Errorf("compareTag(tag, tagSpath) = %d, want 1", got)
		}
		got = compareTag(tag, tagLhash)
		if got != -1 {
			t.Errorf("compareTag(tag, tagLhash) = %d, want -1", got)
		}
		got = compareTag(tag, tagShash)
		if got != 1 {
			t.Errorf("compareTag(tag, tagShash) = %d, want 1", got)
		}
		got = compareTag(tag, tagLtime)
		if got != -1 {
			t.Errorf("compareTag(tag, tagLtime) = %d, want -1", got)
		}
		got = compareTag(tag, tagStime)
		if got != 1 {
			t.Errorf("compareTag(tag, tagStime) = %d, want 1", got)
		}
		got = compareTag(tag, tagLhashStime)
		if got != -1 {
			t.Errorf("compareTag(tag, tagLhashStime) = %d, want -1", got)
		}
		got = compareTag(tag, tagStimeLpath)
		if got != 1 {
			t.Errorf("compareTag(tag, tagStimeLpath) = %d, want 1", got)
		}

		// comparing ... tagShash < tagStime < tagStimeLpath < tagSpath < tag < tagLpath < tagLtime < tagLhashStime < tagLhash
		got = compareTag(tagShash, tagStime)
		if got != -1 {
			t.Errorf("compareTag(tagShash, tagStime) = %d, want -1", got)
		}
		got = compareTag(tagStime, tagStimeLpath)
		if got != -1 {
			t.Errorf("compareTag(tagStime, tagStimeLpath) = %d, want -1", got)
		}
		got = compareTag(tagStimeLpath, tagSpath)
		if got != -1 {
			t.Errorf("compareTag(tagStimeLpath, tagSpath) = %d, want -1", got)
		}
		got = compareTag(tagSpath, tag)
		if got != -1 {
			t.Errorf("compareTag(tagSpath, tag) = %d, want -1", got)
		}
		got = compareTag(tag, tagLpath)
		if got != -1 {
			t.Errorf("compareTag(tag, tagLpath) = %d, want -1", got)
		}
		got = compareTag(tagLpath, tagLtime)
		if got != -1 {
			t.Errorf("compareTag(tagLpath, tagLtime) = %d, want -1", got)
		}
		got = compareTag(tagLtime, tagLhashStime)
		if got != -1 {
			t.Errorf("compareTag(tagLtime, tagLhashStime) = %d, want -1", got)
		}
		got = compareTag(tagLhashStime, tagLhash)
		if got != -1 {
			t.Errorf("compareTag(tagLhashStime, tagLhash) = %d, want -1", got)
		}

	})

	// func findTagIndex(tags []Tag, comparingTag Tag, flag FindField) int
}

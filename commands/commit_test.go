package commands

import (
	"testing"
)

func TestCommit(t *testing.T) {
	// var timestampNow func() int64
	// stub
	timestampNow = func() int64 { return 0x1234 }

	// mock
	fileOp = &FileOp{
		findFilePaths: func(root string) ([]string, error) {
			return []string{"path0", "path1", "path2", "path3", "path4", "path5"}, nil
		},
		rootDir: func() string {
			return "root"
		},
		hashFile: func(path string) (Hash, error) {
			switch path {
			case "root/path3":
				return hashing("a888888888888888888888888888888888888888888888888888888888888883"), nil
			case "root/path2":
				return hashing("b888888888888888888888888888888888888888888888888888888888888882"), nil
			case "root/path1":
				return hashing("c888888888888888888888888888888888888888888888888888888888888881"), nil
			case "root/path4":
				return hashing("d888888888888888888888888888888888888888888888888888888888888884"), nil
			case "root/path5":
				return hashing("e888888888888888888888888888888888888888888888888888888888888885"), nil
			case "root/path0":
				return hashing("f888888888888888888888888888888888888888888888888888888888888880"), nil
			default:
				panic("fileOp.hashFile is called with unknown path")
			}
		},
		timestampFile: func(path string) (int64, error) {
			switch path {
			case "root/path3":
				return 0x0002, nil
			case "root/path2":
				return 0x0003, nil
			case "root/path1":
				return 0x0004, nil
			case "root/path4":
				return 0x0001, nil
			case "root/path5":
				return 0x0000, nil
			case "root/path0":
				return 0x0005, nil
			default:
				panic("fileOp.timestampFile is called with unknown path")
			}
		},
	}

	// func createCommitStructure(fastly bool) (Commit, error)
	t.Run("createCommitStructure(fastly bool) (Commit, error)", func(t *testing.T) {
		got, err := createCommitStructure(false)
		if err != nil {
			t.Errorf("createCommitStructure() return error, %s", err)
		}
		if got.Timestamp != 0x1234 {
			t.Errorf("createCommitStructure() return commit, commit.Timestamp = 0x%.8x, want 0x00001234", got.Timestamp)
		}
		if got.Id != "00001234-f6d531a00f7021b7ca596dc89d9d1e34510d66925aacf9401b23950a47542a41" {
			t.Errorf("createCommitStructure() return commit, commit.Id = %s, want \"00001234-f6d531a00f7021b7ca596dc89d9d1e34510d66925aacf9401b23950a47542a41\"", got.Id)

		}
		if got.Hash.String() != "f6d531a00f7021b7ca596dc89d9d1e34510d66925aacf9401b23950a47542a41" {
			t.Errorf("createCommitStructure() return commit, commit.Hash.String() = %s, want \"f6d531a00f7021b7ca596dc89d9d1e34510d66925aacf9401b23950a47542a41\"", got.Hash.String())
		}

		want := "a888888888888888888888888888888888888888888888888888888888888883 path3"
		if got.Tags[0].String() != want {
			t.Errorf("createCommitStructure() return commit, commit.Tags[0].String() = %s, want \"%s\"", got.Tags[0].String(), want)
		}
		want = "b888888888888888888888888888888888888888888888888888888888888882 path2"
		if got.Tags[1].String() != want {
			t.Errorf("createCommitStructure() return commit, commit.Tags[1].String() = %s, want \"%s\"", got.Tags[1].String(), want)
		}
		want = "c888888888888888888888888888888888888888888888888888888888888881 path1"
		if got.Tags[2].String() != want {
			t.Errorf("createCommitStructure() return commit, commit.Tags[2].String() = %s, want \"%s\"", got.Tags[2].String(), want)
		}
		want = "d888888888888888888888888888888888888888888888888888888888888884 path4"
		if got.Tags[3].String() != want {
			t.Errorf("createCommitStructure() return commit, commit.Tags[3].String() = %s, want \"%s\"", got.Tags[3].String(), want)
		}
		want = "e888888888888888888888888888888888888888888888888888888888888885 path5"
		if got.Tags[4].String() != want {
			t.Errorf("createCommitStructure() return commit, commit.Tags[4].String() = %s, want \"%s\"", got.Tags[4].String(), want)
		}
		want = "f888888888888888888888888888888888888888888888888888888888888880 path0"
		if got.Tags[5].String() != want {
			t.Errorf("createCommitStructure() return commit, commit.Tags[5].String() = %s, want \"%s\"", got.Tags[5].String(), want)
		}
		// FIXME: Add a test case createCommitStructure(true)
	})

	// func tagging(root, relativePath string, withHashing bool) (Tag, error)
	// tagging() is called in createCommitStructure()
}

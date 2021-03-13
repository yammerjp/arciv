package commands

import (
	"testing"
)

func TestRepository(t *testing.T) {
	repo := Repository{Name: "repo_name", Location: RepositoryLocationFile{Path: "root"}}

	// func (repository Repository) String() string
	t.Run("Repository.String()", func(t *testing.T) {
		got := repo.String()
		if got != "name:repo_name type:file path:root" {
			t.Errorf("Repository.String() = %s, want \"name:repo_name type:file path:root\"", got)
		}
	})

	// func (repository Repository) WriteTimeline(timeline []string) error
	// use fileOp.writeLines()
	t.Run("Repository.WriteTimeline()", func(t *testing.T) {
		fileOp = &FileOp{
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/timeline" ||
					len(lines) != 3 ||
					lines[0] != "00000000-0000000000000000000000000000000000000000000000000000000000000000" ||
					lines[1] != "11111111-1111111111111111111111111111111111111111111111111111111111111111" ||
					lines[2] != "22222222-2222222222222222222222222222222222222222222222222222222222222222" {
					t.Errorf("fileOp.writeLines is called with unknown arguments (%s, %s)", path, lines)
				}
				return nil
			},
		}

		err := repo.WriteTimeline([]string{
			"00000000-0000000000000000000000000000000000000000000000000000000000000000",
			"11111111-1111111111111111111111111111111111111111111111111111111111111111",
			"22222222-2222222222222222222222222222222222222222222222222222222222222222",
		})
		if err != nil {
			t.Errorf("Repository.WriteTimeline() return an error \"%s\", want nil", err)
		}
	})

	// func (repository Repository) LoadTimeline() ([]string, error)
	// use fileOp.loadLines()
	t.Run("Repository.LoadTimeline()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				if path != "root/.arciv/timeline" {
					t.Errorf("fileOp.loadLines is called with unkwnown path %s", path)
				}
				return []string{
					"00000000-0000000000000000000000000000000000000000000000000000000000000000",
					"11111111-1111111111111111111111111111111111111111111111111111111111111111",
				}, nil
			},
		}

		got, err := repo.LoadTimeline()
		if err != nil {
			t.Errorf("Repository.LoadTimeline() return an error \"%s\", want nil", err)
		}
		if len(got) != 2 ||
			got[0] != "00000000-0000000000000000000000000000000000000000000000000000000000000000" ||
			got[1] != "11111111-1111111111111111111111111111111111111111111111111111111111111111" {
			t.Errorf("Repository.LoadTimeline() return unexpected values %s, want [\"00000000-0000000000000000000000000000000000000000000000000000000000000000\", \"11111111-1111111111111111111111111111111111111111111111111111111111111111\"]", got)
		}
	})

	// func (repository Repository) LoadLatestCommitId() (string, error)
	t.Run("Repository.LoadLatestCommitId()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				if path != "root/.arciv/timeline" {
					t.Errorf("fileOp.loadLines is called with unkwnown path %s", path)
				}
				return []string{
					"00000000-0000000000000000000000000000000000000000000000000000000000000000",
					"11111111-1111111111111111111111111111111111111111111111111111111111111111",
				}, nil
			},
		}
		got, err := repo.LoadLatestCommitId()
		if err != nil {
			t.Errorf("Repository.LoadLatestCommitId() return an error \"%s\", want nil", err)
		}
		if got != "11111111-1111111111111111111111111111111111111111111111111111111111111111" {
			t.Errorf("Repository.LoadLatestCommitId() return %s, want \"11111111-1111111111111111111111111111111111111111111111111111111111111111\"", got)
		}

		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				if path != "root/.arciv/timeline" {
					t.Errorf("fileOp.loadLines is called with unknown path %s", path)
				}
				return []string{}, nil
			},
		}
		got, err = repo.LoadLatestCommitId()
		if err.Error() != "Commit does not exists" {
			t.Errorf("Repository.LoadLatestCommitId() return error \"%s\", want error \"Commit does not exists\"", err)
		}
	})

	// func (repository Repository) WriteTags(commit Commit, base *Commit) error
	// use fileOp.writeLines()
	t.Run("Repository.WriteTags()", func(t *testing.T) {
		// #arciv-commit-atom
		fileOp = &FileOp{
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/list/aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
					t.Errorf("fileOp.writeLines is called with unknown path %s", path)
				}
				if len(lines) != 4 ||
					lines[0] != "#arciv-commit-atom" ||
					lines[1] != "0000000000000000000000000000000000000000000000000000000000000000 0000/0000" ||
					lines[2] != "1111111111111111111111111111111111111111111111111111111111111111 1111/1111" ||
					lines[3] != "2222222222222222222222222222222222222222222222222222222222222222 2222/2222" {
					t.Errorf("fileOp.WriteLines is called with unknown lines %s", lines)
				}
				return nil
			},
		}
		commit := Commit{
			Id: "aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Tags: []Tag{
				Tag{Path: "0000/0000", Hash: hashing("0000000000000000000000000000000000000000000000000000000000000000"), Timestamp: 0x00000000},
				Tag{Path: "1111/1111", Hash: hashing("1111111111111111111111111111111111111111111111111111111111111111"), Timestamp: 0x11111111},
				Tag{Path: "2222/2222", Hash: hashing("2222222222222222222222222222222222222222222222222222222222222222"), Timestamp: 0x22222222},
			},
		}
		err := repo.WriteTags(commit, nil)
		if err != nil {
			t.Errorf("Repository.WriteTags() return error \"%s\", want nil", err)
		}

		// #arciv-commit-extension from:...
		fileOp = &FileOp{
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/list/aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
					t.Errorf("fileOp.writeLines is called with unknown path %s", path)
				}
				if len(lines) != 2 ||
					lines[0] != "#arciv-commit-extension from:bbbbbbbb-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb" ||
					lines[1] != "+ 1111111111111111111111111111111111111111111111111111111111111111 1111/1111" {
					t.Errorf("fileOp.WriteLines is called with unknown lines %s", lines)
				}
				return nil
			},
		}
		base := Commit{
			Id: "bbbbbbbb-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			Tags: []Tag{
				Tag{Path: "0000/0000", Hash: hashing("0000000000000000000000000000000000000000000000000000000000000000"), Timestamp: 0x00000000},
				Tag{Path: "2222/2222", Hash: hashing("2222222222222222222222222222222222222222222222222222222222222222"), Timestamp: 0x22222222},
			},
		}

		err = repo.WriteTags(commit, &base)
		if err != nil {
			t.Errorf("Repository.WriteTags() return error \"%s\", want nil", err)
		}
	})

	fileOp = &FileOp{
		loadLines: func(path string) ([]string, error) {
			if path == "root/.arciv/timestamps" {
				return []string{
					"#arciv-timestamps of:eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
					"0000000000000000000000000000000000000000000000000000000000000000 00000000",
					"1111111111111111111111111111111111111111111111111111111111111111 11111111",
					"2222222222222222222222222222222222222222222222222222222222222222 22222222",
					"3333333333333333333333333333333333333333333333333333333333333333 33333333",
					"4444444444444444444444444444444444444444444444444444444444444444 44444444",
					"5555555555555555555555555555555555555555555555555555555555555555 55555555",
					"6666666666666666666666666666666666666666666666666666666666666666 66666666",
				}, nil
			}
			if path == "root/.arciv/timeline" {
				return []string{
					"aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					"bbbbbbbb-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
					"cccccccc-cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
					"dddddddd-dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
					"eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
					"ffffffff-ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
				}, nil
			} else if path == "root/.arciv/list/cccccccc-cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" {
				return []string{
					"#arciv-commit-atom",
					"0000000000000000000000000000000000000000000000000000000000000000 0000/0000",
					"1111111111111111111111111111111111111111111111111111111111111111 1111/1111",
					"6666666666666666666666666666666666666666666666666666666666666666 6666/6666",
					"ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff ffff/ffff",
				}, nil
			} else if path == "root/.arciv/list/dddddddd-dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd" {
				return []string{
					"#arciv-commit-extension from:cccccccc-cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
					"+ 2222222222222222222222222222222222222222222222222222222222222222 2222/2222",
					"+ 5555555555555555555555555555555555555555555555555555555555555555 5555/5555",
				}, nil
			} else if path == "root/.arciv/list/eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" {
				return []string{
					"#arciv-commit-extension from:dddddddd-dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
					"- ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff ffff/ffff",
					"+ 3333333333333333333333333333333333333333333333333333333333333333 3333/3333",
					"+ 4444444444444444444444444444444444444444444444444444444444444444 4444/4444",
				}, nil
			} else {
				panic("fileOp.loadLines is called with unknown path: " + path)
			}
		},
	}
	// func (repository Repository) LoadTags(commitId string) (tags []Tag, err error)
	// func (repository Repository) LoadTagsFromExtension(baseCommitId string, body []string) ([]Tag, error)
	// use fileOp.loadLines()
	t.Run("Repository.LoadTags()", func(t *testing.T) {
		got, gotDepth, err := repo.LoadTags("eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
		if err != nil {
			t.Errorf("Repository.LoadTags() return an error \"%s\", want nil", err)
		}
		if len(got) != 7 ||
			got[0].Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" || got[0].Timestamp != 0x00000000 || got[0].Path != "0000/0000" ||
			got[1].Hash.String() != "1111111111111111111111111111111111111111111111111111111111111111" || got[1].Timestamp != 0x11111111 || got[1].Path != "1111/1111" ||
			got[2].Hash.String() != "6666666666666666666666666666666666666666666666666666666666666666" || got[2].Timestamp != 0x66666666 || got[2].Path != "6666/6666" ||
			got[3].Hash.String() != "2222222222222222222222222222222222222222222222222222222222222222" || got[3].Timestamp != 0x22222222 || got[3].Path != "2222/2222" ||
			got[4].Hash.String() != "5555555555555555555555555555555555555555555555555555555555555555" || got[4].Timestamp != 0x55555555 || got[4].Path != "5555/5555" ||
			got[5].Hash.String() != "3333333333333333333333333333333333333333333333333333333333333333" || got[5].Timestamp != 0x33333333 || got[5].Path != "3333/3333" ||
			got[6].Hash.String() != "4444444444444444444444444444444444444444444444444444444444444444" || got[6].Timestamp != 0x44444444 || got[6].Path != "4444/4444" {
			t.Errorf("Repository.LoadTags() = %s", got)
		}
		if gotDepth != 2 {
			t.Errorf("Repository.LoadTags() return depth: %d, want 2", gotDepth)
		}
	})

	// func (repository Repository) LoadCommit(commitId string) (Commit, error)
	// use Repository.LoadTags()
	t.Run("Repository.LoadCommit()", func(t *testing.T) {
		got, err := repo.LoadCommit("eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee")
		if err != nil {
			t.Errorf("Repository.LoadCommit() return an error \"%s\", want nil", err)
		}
		if got.Id != "eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" ||
			got.Timestamp != 0xeeeeeeee ||
			got.Hash.String() != "eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee" ||
			got.Depth != 2 ||
			len(got.Tags) != 7 ||
			got.Tags[0].Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" || got.Tags[0].Timestamp != 0x00000000 || got.Tags[0].Path != "0000/0000" ||
			got.Tags[1].Hash.String() != "1111111111111111111111111111111111111111111111111111111111111111" || got.Tags[1].Timestamp != 0x11111111 || got.Tags[1].Path != "1111/1111" ||
			got.Tags[2].Hash.String() != "2222222222222222222222222222222222222222222222222222222222222222" || got.Tags[2].Timestamp != 0x22222222 || got.Tags[2].Path != "2222/2222" ||
			got.Tags[3].Hash.String() != "3333333333333333333333333333333333333333333333333333333333333333" || got.Tags[3].Timestamp != 0x33333333 || got.Tags[3].Path != "3333/3333" ||
			got.Tags[4].Hash.String() != "4444444444444444444444444444444444444444444444444444444444444444" || got.Tags[4].Timestamp != 0x44444444 || got.Tags[4].Path != "4444/4444" ||
			got.Tags[5].Hash.String() != "5555555555555555555555555555555555555555555555555555555555555555" || got.Tags[5].Timestamp != 0x55555555 || got.Tags[5].Path != "5555/5555" ||
			got.Tags[6].Hash.String() != "6666666666666666666666666666666666666666666666666666666666666666" || got.Tags[6].Timestamp != 0x66666666 || got.Tags[6].Path != "6666/6666" {
			t.Errorf("Repository.LoadCommit() return %s", got.Id)
		}
	})

	// func (repository Repository) LoadCommitFromAlias(alias string) (Commit, error)
	// use Repository.LoadTimeline, Repository.LoadCommit(), findCommitId()
	t.Run("Repository.LoadCommitFromAlias", func(t *testing.T) {
		got, err := repo.LoadCommitFromAlias("c")
		if err != nil {
			t.Errorf("Repository.LoadCommitFromAlias() return an error \"%s\", want nil", err)
		}
		if got.Id != "cccccccc-cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" ||
			got.Timestamp != 0xcccccccc ||
			got.Hash.String() != "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc" ||
			len(got.Tags) != 4 ||
			got.Tags[0].Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" || got.Tags[0].UsedTimestamp != false || got.Tags[0].Path != "0000/0000" ||
			got.Tags[1].Hash.String() != "1111111111111111111111111111111111111111111111111111111111111111" || got.Tags[1].UsedTimestamp != false || got.Tags[1].Path != "1111/1111" ||
			got.Tags[2].Hash.String() != "6666666666666666666666666666666666666666666666666666666666666666" || got.Tags[2].UsedTimestamp != false || got.Tags[2].Path != "6666/6666" ||
			got.Tags[3].Hash.String() != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" || got.Tags[3].UsedTimestamp != false || got.Tags[3].Path != "ffff/ffff" {
			t.Errorf("Repository.LoadCommitFromAlias() return %s", got.Id)
		}
	})

	// func (repository Repository) LoadLatestCommit() (Commit, error)
	// use Repository.LoadLatestCommitId(), Repository.LoadCommit()
	t.Run("Repositor.LoadLatestCommit()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				if path == "root/.arciv/timestamps" {
					return []string{
						"#arciv-timestamps of:ffffffff-ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
						"0000000000000000000000000000000000000000000000000000000000000000 00000000",
						"1111111111111111111111111111111111111111111111111111111111111111 11111111",
					}, nil
				}
				if path == "root/.arciv/timeline" {
					return []string{
						"aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
						"bbbbbbbb-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
						"cccccccc-cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
						"dddddddd-dddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddddd",
						"eeeeeeee-eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
						"ffffffff-ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff",
					}, nil
				} else if path == "root/.arciv/list/ffffffff-ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" {
					return []string{
						"#arciv-commit-atom",
						"0000000000000000000000000000000000000000000000000000000000000000 0000/0000",
						"1111111111111111111111111111111111111111111111111111111111111111 1111/1111",
					}, nil
				} else {
					t.Errorf("fileOp.loadLines is called with unknown path \"%s\"", path)
					panic("")
				}
			},
		}
		got, err := repo.LoadLatestCommit()
		if err != nil {
			t.Errorf("Repository.LoadLatestCommit() return an error \"%s\", want nil", err)
		}
		if got.Id != "ffffffff-ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" ||
			got.Timestamp != 0xffffffff ||
			got.Hash.String() != "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff" ||
			len(got.Tags) != 2 ||
			got.Tags[0].Hash.String() != "0000000000000000000000000000000000000000000000000000000000000000" || got.Tags[0].Timestamp != 0x00000000 || got.Tags[0].Path != "0000/0000" ||
			got.Tags[1].Hash.String() != "1111111111111111111111111111111111111111111111111111111111111111" || got.Tags[1].Timestamp != 0x11111111 || got.Tags[1].Path != "1111/1111" {
			t.Errorf("Repository.LoadCommitFromAlias() return %s", got.Id)
		}
	})

	// func (repository Repository) FetchBlobHashes() ([]string, error)
	// use fileOp.findFilePaths()
	t.Run("Repository.FetchBlobHashes()", func(t *testing.T) {
		fileOp = &FileOp{
			findFilePaths: func(root string) ([]string, error) {
				if root != "root/.arciv/blob" {
					t.Errorf("fileOp.findFilePaths is called with unknown root %s", root)
				}
				return []string{
					"0000000000000000000000000000000000000000000000000000000000000000",
					"1111111111111111111111111111111111111111111111111111111111111111",
					"2222222222222222222222222222222222222222222222222222222222222222.downloading",
				}, nil
			},
		}
		got, err := repo.FetchBlobHashes()
		if err != nil {
			t.Errorf("Repository.FetchBlobHashes() return an error \"%s\", want nil", err)
		}
		if len(got) != 2 ||
			got[0] != "0000000000000000000000000000000000000000000000000000000000000000" ||
			got[1] != "1111111111111111111111111111111111111111111111111111111111111111" {
			t.Errorf("Repository.FetchBlobHashes() return %s", got)
		}
	})

	// func (repository Repository) SendLocalBlob(tag Tag) error
	// use fileOp.rootDir(), fileOp.copyFile()
	t.Run("Repository.SendLocalBlobs()", func(t *testing.T) {
		copied0 := false
		copied1 := false
		fileOp = &FileOp{
			rootDir: func() string { return "local_root" },
			copyFile: func(from, to string) error {
				if from == "local_root/0000/0000" && to == "root/.arciv/blob/0000000000000000000000000000000000000000000000000000000000000000" {
					copied0 = true
					return nil
				}
				if from == "local_root/1111/1111" && to == "root/.arciv/blob/1111111111111111111111111111111111111111111111111111111111111111" {
					copied1 = true
					return nil
				}
				t.Errorf("fileOp.copyFile is called with unknown arguments, (%s, %s)", from, to)
				return nil
			},
		}
		err := repo.SendLocalBlobs([]Tag{
			Tag{Path: "0000/0000", Hash: hashing("0000000000000000000000000000000000000000000000000000000000000000")},
			Tag{Path: "1111/1111", Hash: hashing("1111111111111111111111111111111111111111111111111111111111111111")},
		})
		if err != nil {
			t.Errorf("Repository.SendLocalBlobs() return an error \"%s\", want nil", err)
		}
		if !copied0 || !copied1 {
			t.Errorf("Repository.SendLocalBlobs() failed to copy files")
		}
	})

	// func (repository Repository) ReceiveRemoteBlob(tag Tag) error
	// use fileOp.rootDir(), fileOp.copyFile()
	t.Run("Repository.ReceiveRemoteBlobs()", func(t *testing.T) {
		copied0 := false
		copied1 := false
		fileOp = &FileOp{
			rootDir: func() string { return "local_root" },
			copyFile: func(from, to string) error {
				if from == "root/.arciv/blob/0000000000000000000000000000000000000000000000000000000000000000" && to == "local_root/.arciv/blob/0000000000000000000000000000000000000000000000000000000000000000" {
					copied0 = true
					return nil
				}
				if from == "root/.arciv/blob/1111111111111111111111111111111111111111111111111111111111111111" && to == "local_root/.arciv/blob/1111111111111111111111111111111111111111111111111111111111111111" {
					copied1 = true
					return nil
				}
				t.Errorf("fileOp.copyFile is called with unknown arguments, (%s, %s)", from, to)
				return nil
			},
		}
		err := repo.ReceiveRemoteBlobs([]Tag{
			Tag{Path: "0000/0000", Hash: hashing("0000000000000000000000000000000000000000000000000000000000000000")},
			Tag{Path: "1111/1111", Hash: hashing("1111111111111111111111111111111111111111111111111111111111111111")},
		})
		if err != nil {
			t.Errorf("Repository.ReceiveRemoteBlobs() return an error \"%s\", want nil", err)
		}
		if !copied0 || !copied1 {
			t.Errorf("Repository.ReceiveRemoteBlobs() failed to copy files")
		}
	})

	// func findCommitId(alias string, commitIds []string) (foundCId string, err error)
	fileOp = nil
	t.Run("findCommitId()", func(t *testing.T) {
		commitIds := []string{
			"00000000-0000000000000000000000000000000000000000000000000000000000000000",
			"10000000-0123400000000000000000000000000000000000000000000000000000000000",
			"10001111-0123456789012345678900000000000000000000000000000000000000000000",
			"11111111-1111111111111111111111111111111111111111111111111111111111111111",
		}
		got, err := findCommitId("00", commitIds)
		if err != nil {
			t.Errorf("findCommitId() return an error \"%s\"", err)
		}
		if got != "00000000-0000000000000000000000000000000000000000000000000000000000000000" {
			t.Errorf("findCommitId() = \"%s\", want \"00000000-0000000000000000000000000000000000000000000000000000000000000000\"", got)
		}
		got, err = findCommitId("10000", commitIds)
		if err != nil {
			t.Errorf("findCommitId() return an error \"%s\"", err)
		}
		if got != "10000000-0123400000000000000000000000000000000000000000000000000000000000" {
			t.Errorf("findCommitId() = \"%s\", want \"10000000-0123400000000000000000000000000000000000000000000000000000000000\"", got)
		}

		_, err = findCommitId("01234", commitIds)
		if err.Error() != "The alias refer to more than 1 commit" {
			t.Errorf("findCommitId() return an error \"%s\", want the error \"The alias refer to more than 1 commit\"", err)
		}
		got, err = findCommitId("012345", commitIds)
		if err != nil {
			t.Errorf("findCommitId() return an error \"%s\"", err)
		}
		if got != "10001111-0123456789012345678900000000000000000000000000000000000000000000" {
			t.Errorf("findCommitId() = \"%s\", want \"10001111-0123456789012345678900000000000000000000000000000000000000000000\"", got)
		}
	})

	// func SelfRepo() Repository

	// func (repository Repository) AddCommit(commit Commit) error
	//   use Repository.LoadTimeline(), Repository.LoadCommit(), Repository.WriteTags(), Repository.WriteTimeline()
	//   use fileOp.loadLines(), fileOp.writeLines()
	// append commit
	t.Run("Repository.AddCommit()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				switch path {
				case "root/.arciv/timestamps":
					return []string{"#arciv-timestamps of:11111111-1111111111111111111111111111111111111111111111111111111111111111", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa aaaaaaaa"}, nil
				case "root/.arciv/timeline":
					return []string{
						"00000000-0000000000000000000000000000000000000000000000000000000000000000",
						"11111111-1111111111111111111111111111111111111111111111111111111111111111",
					}, nil
				case "root/.arciv/list/11111111-1111111111111111111111111111111111111111111111111111111111111111":
					return []string{
						"#arciv-commit-atom",
						"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa aaaa/aaaa",
					}, nil
				default:
					t.Errorf("fileOp.loadLines is called with unknown path \"%s\"", path)
					panic("")
				}
			},
			writeLines: func(path string, lines []string) error {
				switch path {
				case "root/.arciv/timeline":
					if len(lines) != 3 ||
						lines[0] != "00000000-0000000000000000000000000000000000000000000000000000000000000000" ||
						lines[1] != "11111111-1111111111111111111111111111111111111111111111111111111111111111" ||
						lines[2] != "22222222-2222222222222222222222222222222222222222222222222222222222222222" {
						t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					}
					return nil
				case "root/.arciv/list/22222222-2222222222222222222222222222222222222222222222222222222222222222":
					if len(lines) != 2 ||
						lines[0] != "#arciv-commit-extension from:11111111-1111111111111111111111111111111111111111111111111111111111111111" ||
						lines[1] != "+ bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb bbbb/bbbb" {
						t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					}
					return nil
				default:
					t.Errorf("fileOp.writeLines is called with unknown path \"%s\"", path)
					panic("")
				}
			},
		}
		err := repo.AddCommit(Commit{
			Id: "22222222-2222222222222222222222222222222222222222222222222222222222222222",
			Tags: []Tag{
				Tag{Path: "aaaa/aaaa", Hash: hashing("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), Timestamp: 0xaaaaaaaa},
				Tag{Path: "bbbb/bbbb", Hash: hashing("bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"), Timestamp: 0xbbbbbbbb},
			},
		})
		if err != nil {
			t.Errorf("Repository.AddCommit() return an error %s, want nil", err)
		}
		// initial commit
		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				switch path {
				case "root/.arciv/timeline":
					return []string{}, nil
				default:
					t.Errorf("fileOp.loadLines is called with unknown path \"%s\"", path)
					panic("")
				}
			},
			writeLines: func(path string, lines []string) error {
				switch path {
				case "root/.arciv/timeline":
					if len(lines) != 1 ||
						lines[0] != "00000000-0000000000000000000000000000000000000000000000000000000000000000" {
						t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					}
					return nil
				case "root/.arciv/list/00000000-0000000000000000000000000000000000000000000000000000000000000000":
					if len(lines) != 2 ||
						lines[0] != "#arciv-commit-atom" ||
						lines[1] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa aaaa/aaaa" {
						t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					}
					return nil
				default:
					t.Errorf("fileOp.writeLines is called with unknow path \"%s\", want nil", err)
					panic("")
				}
			},
		}
		err = repo.AddCommit(Commit{
			Id: "00000000-0000000000000000000000000000000000000000000000000000000000000000",
			Tags: []Tag{
				Tag{Path: "aaaa/aaaa", Hash: hashing("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), Timestamp: 0xaaaaaaaa},
			},
		})
		if err != nil {
			t.Errorf("Repository.AddCommit() return an error %s, want nil", err)
		}

		// commit with depth >= COMMIT_EXTENSION_DEPTH_MAX
		fileOp = &FileOp{
			loadLines: func(path string) ([]string, error) {
				switch path {
				case "root/.arciv/timestamps":
					return []string{}, nil
				case "root/.arciv/timeline":
					return []string{
						"00000000-0000000000000000000000000000000000000000000000000000000000000000",
						"11111111-1111111111111111111111111111111111111111111111111111111111111111",
						"22222222-2222222222222222222222222222222222222222222222222222222222222222",
						"33333333-3333333333333333333333333333333333333333333333333333333333333333",
						"44444444-4444444444444444444444444444444444444444444444444444444444444444",
						"55555555-5555555555555555555555555555555555555555555555555555555555555555",
						"66666666-6666666666666666666666666666666666666666666666666666666666666666",
						"77777777-7777777777777777777777777777777777777777777777777777777777777777",
						"88888888-8888888888888888888888888888888888888888888888888888888888888888",
						"99999999-9999999999999999999999999999999999999999999999999999999999999999",
					}, nil
				case "root/.arciv/list/00000000-0000000000000000000000000000000000000000000000000000000000000000":
					return []string{
						"#arciv-commit-atom",
						"0000000000000000000000000000000000000000000000000000000000000000 0000/0000",
					}, nil
				case "root/.arciv/list/11111111-1111111111111111111111111111111111111111111111111111111111111111":
					return []string{
						"#arciv-commit-extension from:00000000-0000000000000000000000000000000000000000000000000000000000000000",
						"+ 1111111111111111111111111111111111111111111111111111111111111111 1111/1111",
					}, nil
				case "root/.arciv/list/22222222-2222222222222222222222222222222222222222222222222222222222222222":
					return []string{
						"#arciv-commit-extension from:11111111-1111111111111111111111111111111111111111111111111111111111111111",
						"+ 2222222222222222222222222222222222222222222222222222222222222222 2222/2222",
					}, nil
				case "root/.arciv/list/33333333-3333333333333333333333333333333333333333333333333333333333333333":
					return []string{
						"#arciv-commit-extension from:22222222-2222222222222222222222222222222222222222222222222222222222222222",
						"+ 3333333333333333333333333333333333333333333333333333333333333333 3333/3333",
					}, nil
				case "root/.arciv/list/44444444-4444444444444444444444444444444444444444444444444444444444444444":
					return []string{
						"#arciv-commit-extension from:33333333-3333333333333333333333333333333333333333333333333333333333333333",
						"+ 4444444444444444444444444444444444444444444444444444444444444444 4444/4444",
					}, nil
				case "root/.arciv/list/55555555-5555555555555555555555555555555555555555555555555555555555555555":
					return []string{
						"#arciv-commit-extension from:44444444-4444444444444444444444444444444444444444444444444444444444444444",
						"+ 5555555555555555555555555555555555555555555555555555555555555555 5555/5555",
					}, nil
				case "root/.arciv/list/66666666-6666666666666666666666666666666666666666666666666666666666666666":
					return []string{
						"#arciv-commit-extension from:55555555-5555555555555555555555555555555555555555555555555555555555555555",
						"+ 6666666666666666666666666666666666666666666666666666666666666666 6666/6666",
					}, nil
				case "root/.arciv/list/77777777-7777777777777777777777777777777777777777777777777777777777777777":
					return []string{
						"#arciv-commit-extension from:66666666-6666666666666666666666666666666666666666666666666666666666666666",
						"+ 7777777777777777777777777777777777777777777777777777777777777777 7777/7777",
					}, nil
				case "root/.arciv/list/88888888-8888888888888888888888888888888888888888888888888888888888888888":
					return []string{
						"#arciv-commit-extension from:77777777-7777777777777777777777777777777777777777777777777777777777777777",
						"+ 8888888888888888888888888888888888888888888888888888888888888888 8888/8888",
					}, nil
				case "root/.arciv/list/99999999-9999999999999999999999999999999999999999999999999999999999999999":
					return []string{
						"#arciv-commit-extension from:88888888-8888888888888888888888888888888888888888888888888888888888888888",
						"+ 9999999999999999999999999999999999999999999999999999999999999999 9999/9999",
					}, nil
				default:
					t.Errorf("fileOp.loadLines is called with unknown path \"%s\"", path)
					panic("")
				}
			},
			writeLines: func(path string, lines []string) error {
				switch path {
				case "root/.arciv/timeline":
					if len(lines) != 11 ||
						lines[0] != "00000000-0000000000000000000000000000000000000000000000000000000000000000" ||
						lines[1] != "11111111-1111111111111111111111111111111111111111111111111111111111111111" ||
						lines[2] != "22222222-2222222222222222222222222222222222222222222222222222222222222222" ||
						lines[3] != "33333333-3333333333333333333333333333333333333333333333333333333333333333" ||
						lines[4] != "44444444-4444444444444444444444444444444444444444444444444444444444444444" ||
						lines[5] != "55555555-5555555555555555555555555555555555555555555555555555555555555555" ||
						lines[6] != "66666666-6666666666666666666666666666666666666666666666666666666666666666" ||
						lines[7] != "77777777-7777777777777777777777777777777777777777777777777777777777777777" ||
						lines[8] != "88888888-8888888888888888888888888888888888888888888888888888888888888888" ||
						lines[9] != "99999999-9999999999999999999999999999999999999999999999999999999999999999" ||
						lines[10] != "aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
						t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					}
					return nil
				case "root/.arciv/list/aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa":
					if len(lines) != 12 ||
						lines[0] != "#arciv-commit-atom" ||
						lines[1] != "0000000000000000000000000000000000000000000000000000000000000000 0000/0000" ||
						lines[2] != "1111111111111111111111111111111111111111111111111111111111111111 1111/1111" ||
						lines[3] != "2222222222222222222222222222222222222222222222222222222222222222 2222/2222" ||
						lines[4] != "3333333333333333333333333333333333333333333333333333333333333333 3333/3333" ||
						lines[5] != "4444444444444444444444444444444444444444444444444444444444444444 4444/4444" ||
						lines[6] != "5555555555555555555555555555555555555555555555555555555555555555 5555/5555" ||
						lines[7] != "6666666666666666666666666666666666666666666666666666666666666666 6666/6666" ||
						lines[8] != "7777777777777777777777777777777777777777777777777777777777777777 7777/7777" ||
						lines[9] != "8888888888888888888888888888888888888888888888888888888888888888 8888/8888" ||
						lines[10] != "9999999999999999999999999999999999999999999999999999999999999999 9999/9999" ||
						lines[11] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa aaaa/aaaa" {
						t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					}
					return nil
				default:
					t.Errorf("fileOp.writeLines is called with unknow path \"%s\", want nil", err)
					panic("")
				}
			},
		}
		err = repo.AddCommit(Commit{
			Id: "aaaaaaaa-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Tags: []Tag{
				Tag{Path: "0000/0000", Hash: hashing("0000000000000000000000000000000000000000000000000000000000000000"), Timestamp: 0x00000000},
				Tag{Path: "1111/1111", Hash: hashing("1111111111111111111111111111111111111111111111111111111111111111"), Timestamp: 0x11111111},
				Tag{Path: "2222/2222", Hash: hashing("2222222222222222222222222222222222222222222222222222222222222222"), Timestamp: 0x22222222},
				Tag{Path: "3333/3333", Hash: hashing("3333333333333333333333333333333333333333333333333333333333333333"), Timestamp: 0x33333333},
				Tag{Path: "4444/4444", Hash: hashing("4444444444444444444444444444444444444444444444444444444444444444"), Timestamp: 0x44444444},
				Tag{Path: "5555/5555", Hash: hashing("5555555555555555555555555555555555555555555555555555555555555555"), Timestamp: 0x55555555},
				Tag{Path: "6666/6666", Hash: hashing("6666666666666666666666666666666666666666666666666666666666666666"), Timestamp: 0x66666666},
				Tag{Path: "7777/7777", Hash: hashing("7777777777777777777777777777777777777777777777777777777777777777"), Timestamp: 0x77777777},
				Tag{Path: "8888/8888", Hash: hashing("8888888888888888888888888888888888888888888888888888888888888888"), Timestamp: 0x88888888},
				Tag{Path: "9999/9999", Hash: hashing("9999999999999999999999999999999999999999999999999999999999999999"), Timestamp: 0x99999999},
				Tag{Path: "aaaa/aaaa", Hash: hashing("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"), Timestamp: 0xaaaaaaaa},
			},
		})
		if err != nil {
			t.Errorf("Repository.AddCommit() return an error %s, want nil", err)
		}
	})
}

package commands

import (
	"testing"
)

func TestRepositoryInit(t *testing.T) {
	repo := Repository{Name: "repo-name", Path: "root", PathType: PATH_FILE}
	mkdirAll := func(path string) error {
		switch path {
		case "root/.arciv/blob":
			return nil
		case "root/.arciv/list":
			return nil
		default:
			panic("fileOp.mkdirAll is called with unknown path " + path)
		}
	}

	// func (repository Repository) Init() error
	t.Run("Repository.Init()", func(t *testing.T) {
		// do not create .arciv/repositories, .arciv/timeline
		fileOp = &FileOp{
			mkdirAll: mkdirAll,
			writeLines: func(_ string, _ []string) error {
				panic("fileOp.writeLines is called")
			},
			findFilePaths: func(root string) ([]string, error) {
				if root != "root/.arciv" {
					panic("fileOp.findFilePaths is called with a unknown argument " + root)
				}
				return []string{".arciv/blob/00000000-0000000000000000000000000000000000000000000000000000000000000000", ".arciv/blob/11111111-1111111111111111111111111111111111111111111111111111111111111111", ".arciv/repositories", ".arciv/timeline"}, nil
			},
		}
		err := repo.Init()
		if err != nil {
			t.Errorf("Repository.Init() return a error \"%s\"", err)
		}

		// create .arciv/repositories, .arciv/timeline
		fileIsCreatedRepositories := false
		fileIsCreatedTimeline := false
		fileOp = &FileOp{
			mkdirAll: mkdirAll,
			writeLines: func(path string, lines []string) error {
				if len(lines) != 0 {
					panic("fileOp.writeLines is called with unknown lines")
				}
				switch path {
				case "root/.arciv/repositories":
					fileIsCreatedRepositories = true
					return nil
				case "root/.arciv/timeline":
					fileIsCreatedTimeline = true
					return nil
				default:
					panic("fileOp.mkdirAll is called with unknown path " + path)
				}
			},
			findFilePaths: func(root string) ([]string, error) {
				if root != "root/.arciv" {
					panic("fileOp.findFilePaths is called with a unknown argument " + root)
				}
				return []string{}, nil
			},
		}
		err = repo.Init()
		if err != nil {
			t.Errorf("Repository.Init() return a error \"%s\"", err)
		}
		if !fileIsCreatedRepositories {
			t.Errorf("Repository.Init() does not create '.arciv/repositories'")
		}
		if !fileIsCreatedTimeline {
			t.Errorf("Repository.Init() does not create '.arciv/timeline'")
		}
	})
}

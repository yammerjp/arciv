package commands

import (
	"testing"
)

func TestRepositoryInit(t *testing.T) {
	mkdirAll := func(path string) error {
		switch path {
		case "root/.arciv/blob":
			return nil
		case "root/.arciv/list":
			return nil
		case "root/.arciv/restore-request":
			return nil
		default:
			panic("fileOp.mkdirAll is called with unknown path " + path)
		}
	}

	// func (repository Repository) Init() error
	t.Run("Repository.Init() (type:file)", func(t *testing.T) {
		repo := Repository{Name: "repo-name", Location: RepositoryLocationFile{Path: "root"}}
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
				return []string{"blob/00000000-0000000000000000000000000000000000000000000000000000000000000000", "blob/11111111-1111111111111111111111111111111111111111111111111111111111111111", "repositories", "timeline", "timestamps"}, nil
			},
		}
		err := repo.Init()
		if err != nil {
			t.Errorf("Repository.Init() return a error \"%s\"", err)
		}

		// create .arciv/repositories, .arciv/timeline
		fileIsCreatedRepositories := false
		fileIsCreatedTimeline := false
		fileIsCreatedTimeStamps := false
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
				case "root/.arciv/timestamps":
					fileIsCreatedTimeStamps = true
					return nil
				default:
					panic("fileOp.writeLines is called with unknown path " + path)
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
		if !fileIsCreatedTimeStamps {
			t.Errorf("Repository.Init() does not create '.arciv/timestamps'")
		}
	})

	t.Run("Repository.Init() (type:s3)", func(t *testing.T) {
		s3Op = &S3Op{
			findFilePaths: func(region, bucket, root string) ([]string, error) {
				if region != "region-name" {
					t.Errorf("s3Op.findFilePaths() gets invalid region name, %s", region)
				}
				if bucket != "bucket-name" {
					t.Errorf("s3Op.findFilePaths() gets invalid bucket name, %s", bucket)
				}
				if root != ".arciv" {
					t.Errorf("s3Op.findFilePaths() gets invalid root, %s", root)
				}
				return []string{"repositories", "timeline", "timestamps", "blob/0000000000000000000000000000000000000000000000000000000000000000"}, nil
			},
		}
		repo := Repository{Name: "repo-name", Location: RepositoryLocationS3{BucketName: "bucket-name", RegionName: "region-name"}}
		err := repo.Init()
		if err != nil {
			t.Errorf("Repository.Init() return a error \"%s\"", err)
		}
	})

	t.Run("Repository.Init() (type:s3)", func(t *testing.T) {
		fileIsCreatedRepositories := false
		fileIsCreatedTimeline := false
		fileIsCreatedTimeStamps := false
		s3Op = &S3Op{
			findFilePaths: func(region, bucket, root string) ([]string, error) {
				if region != "region-name" {
					t.Errorf("s3Op.findFilePaths() gets invalid region name, %s", region)
				}
				if bucket != "bucket-name" {
					t.Errorf("s3Op.findFilePaths() gets invalid bucket name, %s", bucket)
				}
				if root != ".arciv" {
					t.Errorf("s3Op.findFilePaths() gets invalid root, %s", root)
				}
				return []string{}, nil
			},
			writeLines: func(region, bucket, path string, lines []string) error {
				if region != "region-name" {
					t.Errorf("s3Op.writeLines() gets invalid region name, %s", region)
				}
				if bucket != "bucket-name" {
					t.Errorf("s3Op.writeLines() gets invalid bucket name, %s", bucket)
				}
				if path == ".arciv/repositories" {
					fileIsCreatedRepositories = true
				} else if path == ".arciv/timeline" {
					fileIsCreatedTimeline = true
				} else if path == ".arciv/timestamps" {
					fileIsCreatedTimeStamps = true
				} else {
					t.Errorf("s3Op.writeLines() gets invalid path, %s", path)
				}
				if len(lines) != 0 {
					t.Errorf("s3Op.writeLines() gets invalid lines, %s", lines)
				}
				return nil
			},
		}
		repo := Repository{Name: "repo-name", Location: RepositoryLocationS3{BucketName: "bucket-name", RegionName: "region-name"}}
		err := repo.Init()
		if err != nil {
			t.Errorf("Repository.Init() return a error \"%s\"", err)
		}
		if !fileIsCreatedRepositories {
			t.Errorf("Repository.Init() does not create '.arciv/repositories'")
		}
		if !fileIsCreatedTimeline {
			t.Errorf("Repository.Init() does not create '.arciv/timeline'")
		}
		if !fileIsCreatedTimeStamps {
			t.Errorf("Repository.Init() does not create '.arciv/timestamps'")
		}
	})
}

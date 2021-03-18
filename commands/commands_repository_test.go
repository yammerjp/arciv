package commands

import (
	"fmt"
	"os"
	"testing"
)

func TestCommandsRepository(t *testing.T) {

	// func strs2repository(elements []string) (Repository, error)
	t.Run("strs2repository()", func(t *testing.T) {
		got, err := strs2repository([]string{"path:path/to/dir", "name:repoName", "type:file"})
		if err != nil {
		}
		if got.String() != "name:repoName type:file path:path/to/dir" {
			t.Errorf("strs2repository() return Repository{%s}, want Repository{name:repoName type:file path:path/to/dir}", got)
		}
		got, err = strs2repository([]string{"name:repo-s3", "type:s3", "bucket:bucket-name", "region:region-name"})
		if err != nil {
			t.Errorf("strs2repository() return an error \"%s\", want nil", err)
		}
		if got.String() != "name:repo-s3 type:s3 region:region-name bucket:bucket-name" {
			t.Errorf("strs2repository() return Repository{%s}, want Repository{name:repo-s3 type:s3 region:region-name bucket:bucket-name}", got)
		}

		_, err = strs2repository([]string{"name:repo-name path:path/to/dir"})
		if err.Error() != "Unknown repository's type" {
			t.Errorf("strs2repository() return an error \"%s\", want \"Unknown repository's type\"", err)
		}

		_, err = strs2repository([]string{"path:hoge", "type:file"})
		if err.Error() != "Repository's name is not specified" {

			t.Errorf("strs2repository() return an error \"%s\", want \"Repository's name is not specified\"", err)
		}
		_, err = strs2repository([]string{"type:s3", "name:n", "bucket:b"})
		if err.Error() != "Repository's type is s3, but bucket or region is not specified" {
			t.Errorf("strs2repository() return an error \"%s\", want \"Repository's type is s3, but bucket or region is not specified\"", err)
		}
		_, err = strs2repository([]string{"type:s3", "name:n", "region:r"})
		if err.Error() != "Repository's type is s3, but bucket or region is not specified" {
			t.Errorf("strs2repository() return an error \"%s\", want \"Repository's type is s3, but bucket or region is not specified\"", err)
		}
		_, err = strs2repository([]string{"type:s3", "name:a", "region:r", "unknownstring"})
		if err.Error() != "Repository definition is invalid syntax" {
			t.Errorf("strs2repository() return an error \"%s\", want \"Repository definition is invalid syntax\"", err)
		}
	})

	// func writeRepos(repos []Repository) error
	//   use fileOp.writeLines(), fileOp.rootDir()
	t.Run("writeRepos()", func(t *testing.T) {
		fileOp = &FileOp{
			rootDir: func() string {
				return "root"
			},
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/repositories" {
					t.Errorf("fileOp.writeLines is called with unknown path %s", path)
				}
				if len(lines) != 3 ||
					lines[0] != "name:repo-relative type:file path:path-relative" ||
					lines[1] != "name:repo-absolute type:file path:/path-absolute" ||
					lines[2] != "name:repo-new type:file path:repo-new" {
					t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
					for i, line := range lines {
						fmt.Fprintf(os.Stderr, "    lines[%d] : \"%s\"\n", i, line)
					}
				}
				return nil
			},
		}
		err := writeRepos([]Repository{
			Repository{Name: "repo-relative", Location: RepositoryLocationFile{Path: "path-relative"}},
			Repository{Name: "repo-absolute", Location: RepositoryLocationFile{Path: "/path-absolute"}},
			Repository{Name: "repo-new", Location: RepositoryLocationFile{Path: "repo-new"}},
		})
		if err != nil {
			t.Errorf("writeRepos() return an error \"%s\", want nil", err)
		}
	})

	// DI functions used in loadRepos()
	loadLines := func(path string) ([]string, error) {
		if path != "root/.arciv/repositories" {
			panic("fileOp.loadLines is called with unknown path " + path)
		}
		return []string{
			"name:repo-relative type:file path:path-relative",
			"name:repo-absolute type:file path:/path-absolute",
		}, nil
	}
	rootDir := func() string {
		return "root"
	}

	// func loadRepos() ([]Repository, error)
	//   use fileOp.loadLines(), fileOp.rootDir(), SelfRepo(), createRepoStruct()
	t.Run("loadRepos()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: loadLines,
			rootDir:   rootDir,
		}
		repos, err := loadRepos()
		if err != nil {
			t.Errorf("loadRepos() return an error \"%s\", want nil", err)
		}
		if len(repos) != 3 ||
			repos[0].Name != "self" || repos[0].Location.String() != "type:file path:root" ||
			repos[1].Name != "repo-relative" || repos[1].Location.String() != "type:file path:path-relative" ||
			repos[2].Name != "repo-absolute" || repos[2].Location.String() != "type:file path:/path-absolute" {
			t.Errorf("loadRepos() = %s", repos)
		}
	})

	// func findRepo(name string) (Repository, error)
	//   use loadRepos()
	t.Run("findRepo()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: loadLines,
			rootDir:   rootDir,
		}
		repo, err := findRepo("self")
		if err != nil {
			t.Errorf("findRepo(\"self\") return an error \"%s\", want nil", err)
		}
		if repo.String() != "name:self type:file path:root" {
			t.Errorf("findRepo(\"self\") %s, want Repository{name:self type:file path:root}", repo)
		}

		repo, err = findRepo("repo-relative")
		if err != nil {
			t.Errorf("findRepo(\"repo-relative\") return an error \"%s\", want nil", err)
		}
		if repo.String() != "name:repo-relative type:file path:path-relative" {
			t.Errorf("findRepo(\"repo-relative\") %s, want Repository{name:self type:file path:path-relative}", repo)
		}

		repo, err = findRepo("repo-relativ")
		if err.Error() != "Repository is not found" {
			t.Errorf("findRepo(\"repo-relativ\") return an error \"%s\", want \"Repository is not found\"", err)
		}
	})

	// func repositoryActionAdd(name string, url string) error
	//   use loadRepos(), createRepoStruct(), writeRepos(), repo.Init()
	t.Run("repositoryActionAdd()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: loadLines,
			rootDir:   rootDir,
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/repositories" {
					t.Errorf("fileOp.writeLines is called with unknown path %s", path)
				}
				if len(lines) != 3 ||
					lines[0] != "name:repo-relative type:file path:path-relative" ||
					lines[1] != "name:repo-absolute type:file path:/path-absolute" ||
					lines[2] != "name:repo-new type:file path:path-new" {
					t.Errorf("fileOp.writeLines is called with unkwnown lines %s", lines)
				}
				return nil
			},
			// for Repository.Init()
			findFilePaths: func(path string) ([]string, error) {
				if path != "path-new/.arciv" {
					t.Errorf("fileOp.findFilePaths is called with unknown path %s", path)
				}
				return []string{"repositories", "timeline", "timestamps"}, nil
			},
			// for Repository.Init()
			mkdirAll: func(path string) error {
				if path != "path-new/.arciv/list" && path != "path-new/.arciv/blob" && path != "path-new/.arciv/restore-request" {
					t.Errorf("fileOp.mkdirAll is called with unknown path %s", path)
				}
				return nil
			},
		}

		// bad case, repository name includes white space
		err := repositoryActionAdd([]string{"name:repo include white space", "type:file", "path:path-new"})
		if err.Error() != "Including space in repository definition is not supported" {
			t.Errorf("repositoryActionAdd([]string{\"name:repo include white space\", \"type:file\", \"path:path-new\"}) return an error \"%s\", want \"Repository name must not include space\"", err)
		}

		// bad case, repository name conflicts
		err = repositoryActionAdd([]string{"name:repo-relative", "type:file", "path:/path/to/dir"})
		if err.Error() != "The repository name already exists" {
			t.Errorf("repositoryActionAdd([]string{\"name:repo-relative\",\"type:file\", \"path:/path/to/dir\"}) return an error \"%s\", want \"The repository name already exists\"", err)
		}
		// bad case, repository path is invalid
		err = repositoryActionAdd([]string{"name:repo-new", "invalid-path"})
		if err.Error() != "Repository definition is invalid syntax" {
			t.Errorf("repositoryActionAdd(\"repo-new\", \"invalid-path\") return an error \"%s\", want \"Repository path must be file:// of s3:// ...\"", err)
		}

		// success case
		err = repositoryActionAdd([]string{"name:repo-new", "type:file", "path:path-new"})
		if err != nil {
			t.Errorf("repositoryActionAdd([]string{\"name:repo-new\", \"type:file\", \"path:path-new\"}) return an error \"%s\", want nil", err)
		}

		s3Op = &S3Op{
			findFilePaths: func(region string, bucket string, root string) (relativePaths []string, err error) {
				if region != "s3-region" || bucket != "s3-bucket" || root != ".arciv" {
					t.Errorf("s3Op.findFilePaths() is called with unknown arguments: %s, %s, %s", region, bucket, root)
				}
				return []string{"repositories", "timeline", "timestamps"}, nil
			},
		}
		fileOp = &FileOp{
			loadLines: loadLines,
			rootDir:   rootDir,
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/repositories" {
					t.Errorf("fileOp.writeLines is called with unknown path %s", path)
				}
				if len(lines) != 3 ||
					lines[0] != "name:repo-relative type:file path:path-relative" ||
					lines[1] != "name:repo-absolute type:file path:/path-absolute" ||
					lines[2] != "name:repo-s3-new type:s3 region:s3-region bucket:s3-bucket" {
					t.Errorf("fileOp.writeLines is called with unkwnown lines %s", lines)
				}
				return nil
			},
		}

		err = repositoryActionAdd([]string{"type:s3", "name:repo-s3-new", "region:s3-region", "bucket:s3-bucket"})
		if err != nil {
			t.Errorf("repositoryActionAdd([]string{\"type:s3\", \"name:repo-s3-new\", \"region:s3-region\", \"bucket:s3-bucket\"}) return an error \"%s\", want nil", err)
		}
	})

	// func repositoryActionRemove(name string) error
	//   use loadRepos(), writeRepos()
	t.Run("repositoryActionRemove()", func(t *testing.T) {
		fileOp = &FileOp{
			loadLines: loadLines,
			rootDir:   rootDir,
			writeLines: func(path string, lines []string) error {
				if path != "root/.arciv/repositories" {
					t.Errorf("fileOp.writeLines is called with unknown path %s", path)
				}
				if len(lines) != 1 ||
					lines[0] != "name:repo-absolute type:file path:/path-absolute" {
					t.Errorf("fileOp.writeLines is called with unknown lines %s", lines)
				}
				return nil
			},
		}
		// bad case, try to remove self
		err := repositoryActionRemove("self")
		if err.Error() != "self must be exist" {
			t.Errorf("repositoryActionRemove(\"self\") return an error \"%s\", want the error \"self must be exist\"", err)
		}

		// bad case, try to remove a repository which not exists
		err = repositoryActionRemove("repo-not-exists")
		if err.Error() != "The repository is not found" {
			t.Errorf("repositoryActionRemove(\"repo-not-exists\") return an error \"%s\", want the error \"The repository is not found\"", err)
		}

		// success case
		err = repositoryActionRemove("repo-relative")
		if err != nil {
			t.Errorf("repositoryActionRemove(\"repo-relative\") return an error \"%s\", want nil", err)
		}
	})

	// func repositoryAction(args []string) (err error)
	// func repositoryActionShow() error
}

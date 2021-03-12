package commands

import (
	"fmt"
	"os"
	"testing"
)

func TestCommandsRepository(t *testing.T) {
	// func createRepoStruct(name string, url string) (Repository, error)
	t.Run("createRepoStruct()", func(t *testing.T) {
		repo, err := createRepoStruct("repo-name", "file:/invalid-path")
		if err.Error() != "Repository path must be file:// or s3:// ..." {
			t.Errorf("createRepoStruct() return an error \"%s\", want \"Repository path must be file:// of s3:// ...\"", err)
		}
		repo, err = createRepoStruct("repo-name", "file://relative-path")
		if err != nil {
			t.Errorf("createRepoStruct() return an error \"%s\", want nil", err)
		}
		if repo.Name != "repo-name" ||
			repo.Location.String() != "file://relative-path" {
			t.Errorf("createRepoStruct() = %s, want {Name: \"repo-name\", Location: file://relative-path", repo)
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
					lines[0] != "repo-relative file://path-relative" ||
					lines[1] != "repo-absolute file:///path-absolute" ||
					lines[2] != "repo-new file://repo-new" {
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
			"repo-relative file://path-relative",
			"repo-absolute file:///path-absolute",
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
			repos[0].Name != "self" || repos[0].Location.String() != "file://root" ||
			repos[1].Name != "repo-relative" || repos[1].Location.String() != "file://path-relative" ||
			repos[2].Name != "repo-absolute" || repos[2].Location.String() != "file:///path-absolute" {
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
		if repo.Name != "self" || repo.Location.String() != "file://root" {
			t.Errorf("findRepo(\"self\") %s, want Repository{Name: \"self\", Location: file://root", repo)
		}

		repo, err = findRepo("repo-relative")
		if err != nil {
			t.Errorf("findRepo(\"repo-relative\") return an error \"%s\", want nil", err)
		}
		if repo.Name != "repo-relative" || repo.Location.String() != "file://path-relative" {
			t.Errorf("findRepo(\"repo-relative\") %s, want Repository{Name: \"repo-relative\", Location: file://path-relative", repo)
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
					lines[0] != "repo-relative file://path-relative" ||
					lines[1] != "repo-absolute file:///path-absolute" ||
					lines[2] != "repo-new file://path-new" {
					t.Errorf("fileOp.writeLines is called with unkwnown lines %s", lines)
				}
				return nil
			},
			// for Repository.Init()
			findFilePaths: func(path string) ([]string, error) {
				if path != "path-new/.arciv" {
					t.Errorf("fileOp.findFilePaths is called with unknown path %s", path)
				}
				return []string{"repositories", "timeline"}, nil
			},
			// for Repository.Init()
			mkdirAll: func(path string) error {
				if path != "path-new/.arciv/list" && path != "path-new/.arciv/blob" {
					t.Errorf("fileOp.mkdirAll is called with unknown path %s", path)
				}
				return nil
			},
		}

		// bad case, repository name includes white space
		err := repositoryActionAdd("repo include white space", "file://path-new")
		if err.Error() != "Repository name must not include space" {
			t.Errorf("repositoryActionAdd(\"repo include white space\", \"file://path-new\") return an error \"%s\", want \"Repository name must not include space\"", err)
		}

		// bad case, repository name conflicts
		err = repositoryActionAdd("repo-relative", "file:///path/to/dir")
		if err.Error() != "The repository name already exists" {
			t.Errorf("repositoryActionAdd(\"repo-relative\", \"file:///path/to/dir\") return an error \"%s\", want \"The repository name already exists\"", err)
		}
		// bad case, repository path is invalid
		err = repositoryActionAdd("repo-new", "invalid-path")
		if err.Error() != "Repository path must be file:// or s3:// ..." {
			t.Errorf("repositoryActionAdd(\"repo-new\", \"invalid-path\") return an error \"%s\", want \"Repository path must be file:// of s3:// ...\"", err)
		}

		// success case
		err = repositoryActionAdd("repo-new", "file://path-new")
		if err != nil {
			t.Errorf("repositoryActionAdd(\"repo-new\", \"file://path-new\") return an error \"%s\", want nil", err)
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
					lines[0] != "repo-absolute file:///path-absolute" {
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
